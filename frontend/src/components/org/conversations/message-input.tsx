import type { MentionCandidate, MentionQuery } from "./lib/mentions"
import type { ChatMessage } from "./lib/messages"
import type { InfiniteData } from "@tanstack/react-query"

import { useQueryClient } from "@tanstack/react-query"
import { EmojiPicker } from "frimousse"
import { CheckIcon, PaperclipIcon, PencilIcon, SendHorizontalIcon, SmileIcon, XIcon } from "lucide-react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"
import { useThrottledCallback } from "use-debounce"

import { useGetConversationsIdMembers, usePatchConversationsMessagesMessageId } from "@/api/conversations/conversations"
import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { cn } from "@/lib/utils"
import { useChatUi } from "@/stores/chat-ui"

import { AttachmentTray } from "./attachment-tray"
import { useChatWs } from "./chat-provider"
import { detectMention, insertAtCaret, insertMention, resolveMentions } from "./lib/mentions"
import { replaceMessage } from "./lib/optimistic"
import { chatKeys } from "./lib/query-keys"
import { MentionPopover } from "./mention-popover"
import { capFiles, MAX_MEDIA_PER_MESSAGE } from "./upload/upload-manager"
import { useSendAttachments } from "./use-send-attachments"
import { useSendMessage } from "./use-send-message"

type MessagesCache = InfiniteData<ChatMessage[]>

// Cap the mention autocomplete so a huge channel doesn't render a wall of rows.
const MAX_MENTION_ROWS = 8

// Leading-edge throttle for outgoing typing signals — at most once per window,
// fired immediately on the first keystroke of a burst.
const TYPING_THROTTLE_MS = 3000

interface MessageInputProps {
  convId: string
}

/**
 * The thread composer: an auto-growing textarea with @mention autocomplete, an
 * emoji picker, a reply strip, and send. Enter sends / Shift+Enter inserts a
 * newline — but an open mention OR emoji popover GUARDS Enter so it selects /
 * stays put instead of firing an accidental send. Mentions are re-derived from
 * the final text at send time (`resolveMentions`), so deleting an inserted name
 * correctly drops it. The composer owns clearing its own input.
 */
export function MessageInput({ convId }: MessageInputProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { send } = useSendMessage(convId)
  const { sendWithAttachments } = useSendAttachments(convId)
  const editMutation = usePatchConversationsMessagesMessageId()
  const { typing } = useChatWs()

  // Leading-edge throttle: at most one `typing` frame per TYPING_THROTTLE_MS,
  // fired on the first non-empty keystroke of a burst.
  const sendTyping = useThrottledCallback(() => typing(convId), TYPING_THROTTLE_MS, { trailing: false })

  const replyTo = useChatUi((s) => s.replyTo)
  const setReplyTo = useChatUi((s) => s.setReplyTo)
  const editingMessageId = useChatUi((s) => s.editingMessageId)
  const setEditing = useChatUi((s) => s.setEditing)

  const key = chatKeys.messages(convId)

  const { data: membersData } = useGetConversationsIdMembers(convId)
  // Map API members → mention candidates (id + display name), dropping anyone
  // without both. This is the FULL list used to resolve mentions at send time.
  const members: MentionCandidate[] = (membersData?.status === 200 ? (membersData.data.data ?? []) : [])
    .map((m) => ({ id: m.user_id ?? m.user?.id ?? "", name: m.user?.name ?? "" }))
    .filter((m) => m.id && m.name)

  const [value, setValue] = useState("")
  const [mentionQuery, setMentionQuery] = useState<MentionQuery | null>(null)
  const [mentionIndex, setMentionIndex] = useState(0)
  const [emojiOpen, setEmojiOpen] = useState(false)
  // Staged (pre-send) attachments — ephemeral composer state, cleared on send.
  const [files, setFiles] = useState<File[]>([])
  const [dragOver, setDragOver] = useState(false)

  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)
  // Last known caret offset, kept fresh so an emoji inserted while the picker
  // popover holds focus still lands where the user left off.
  const caretPosRef = useRef(0)
  // When set, the value-sync effect restores the caret here after a programmatic
  // edit (mention/emoji insert) and refocuses the textarea.
  const pendingCaretRef = useRef<number | null>(null)

  // Members matching the in-progress token, longest list capped. Case-insensitive
  // prefix-or-substring match keeps it forgiving.
  const mentionMatches: MentionCandidate[] = mentionQuery
    ? members.filter((m) => m.name.toLowerCase().includes(mentionQuery.token.toLowerCase())).slice(0, MAX_MENTION_ROWS)
    : []
  const mentionOpen = mentionMatches.length > 0

  // The referenced message for the reply strip, read live from the message cache.
  const replyMessage = replyTo
    ? queryClient
        .getQueryData<MessagesCache>(key)
        ?.pages.flat()
        .find((m) => m.id === replyTo)
    : undefined

  // The message currently being edited, read live from the cache — drives the
  // "Editing" strip and (via the effect below) prefills the textarea.
  const editingMessage = editingMessageId
    ? queryClient
        .getQueryData<MessagesCache>(key)
        ?.pages.flat()
        .find((m) => m.id === editingMessageId)
    : undefined

  // Auto-grow: reset then match content height; CSS `max-h-40` caps it and lets
  // it scroll past that.
  useEffect(() => {
    const el = textareaRef.current
    if (!el) return
    el.style.height = "auto"
    el.style.height = `${el.scrollHeight}px`
  }, [value])

  // Restore caret + focus after a programmatic edit (mention / emoji insert).
  useEffect(() => {
    const caret = pendingCaretRef.current
    if (caret == null) return
    pendingCaretRef.current = null
    const el = textareaRef.current
    if (!el) return
    el.focus()
    el.setSelectionRange(caret, caret)
    caretPosRef.current = caret
  }, [value])

  // Autofocus on mount and whenever a reply is (re)targeted at this composer.
  useEffect(() => {
    textareaRef.current?.focus()
  }, [replyTo])

  // Enter/exit edit mode. Entering: prefill the textarea with the target's
  // current content, park the caret at the end, and drop any pending reply.
  // Exiting (cancel or success): clear the draft back to empty.
  useEffect(() => {
    if (editingMessageId) {
      const content =
        queryClient
          .getQueryData<MessagesCache>(key)
          ?.pages.flat()
          .find((m) => m.id === editingMessageId)?.content ?? ""
      setValue(content)
      pendingCaretRef.current = content.length
      setReplyTo(null)
      setFiles([])
    } else {
      setValue("")
      caretPosRef.current = 0
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editingMessageId])

  // Recompute the active `@token` from the text up to the caret.
  function refreshMention(nextValue: string, caret: number) {
    caretPosRef.current = caret
    const query = detectMention(nextValue.slice(0, caret))
    setMentionQuery(query)
    setMentionIndex(0)
  }

  function handleChange(e: React.ChangeEvent<HTMLTextAreaElement>) {
    const next = e.target.value
    setValue(next)
    refreshMention(next, e.target.selectionStart)
    if (next.trim()) sendTyping()
  }

  // Caret moved without a value change (click / arrow keys) — re-detect.
  function handleCaretSync(e: React.SyntheticEvent<HTMLTextAreaElement>) {
    refreshMention(value, e.currentTarget.selectionStart)
  }

  // Commit a chosen member: swap the `@token` span for `@<Name> ` and park the
  // caret past the trailing space.
  function commitMention(member: MentionCandidate) {
    if (!mentionQuery) return
    const caret = textareaRef.current?.selectionStart ?? value.length
    const { value: next, caret: nextCaret } = insertMention(value, mentionQuery, caret, member.name)
    setValue(next)
    pendingCaretRef.current = nextCaret
    setMentionQuery(null)
  }

  function insertEmoji(emoji: string) {
    const caret = caretPosRef.current
    const { value: next, caret: nextCaret } = insertAtCaret(value, caret, emoji)
    setValue(next)
    pendingCaretRef.current = nextCaret
    setEmojiOpen(false)
  }

  // Optimistically apply an edit to the cached bubble, then PATCH. On success
  // reconcile with the server copy; on error refetch to restore the truth.
  function submitEdit(messageId: string, content: string) {
    const existing = queryClient
      .getQueryData<MessagesCache>(key)
      ?.pages.flat()
      .find((m) => m.id === messageId)
    if (existing) {
      queryClient.setQueryData<MessagesCache>(key, (old) =>
        replaceMessage(old, { ...existing, content, is_edited: true })
      )
    }
    editMutation.mutate(
      { messageId, data: { content } },
      {
        onSuccess: (res) => {
          const server = res.status === 200 ? (res.data.data as ChatMessage | undefined) : undefined
          if (server) queryClient.setQueryData<MessagesCache>(key, (old) => replaceMessage(old, server))
        },
        onError: () => {
          queryClient.invalidateQueries({ queryKey: key })
        },
      }
    )
    setEditing(null)
  }

  function cancelEdit() {
    setEditing(null)
  }

  // Stage files, capping the combined set at the per-message media limit. When
  // the cap truncates the selection, surface it so the excess isn't dropped
  // silently.
  function addFiles(incoming: File[]) {
    if (incoming.length === 0) return
    const combined = [...files, ...incoming]
    const capped = capFiles(combined)
    const dropped = combined.length - capped.length
    if (dropped > 0) {
      toast.warning(
        t("conversations.attachments.capExceeded", { max: MAX_MEDIA_PER_MESSAGE, count: dropped })
      )
    }
    setFiles(capped)
  }

  function removeFile(index: number) {
    setFiles((prev) => prev.filter((_, i) => i !== index))
  }

  function onFilePick(e: React.ChangeEvent<HTMLInputElement>) {
    if (e.target.files) addFiles(Array.from(e.target.files))
    // Reset so re-picking the same file still fires a change event.
    e.target.value = ""
  }

  // Pasted images (and any other files) land in the tray; text paste is untouched.
  function handlePaste(e: React.ClipboardEvent<HTMLTextAreaElement>) {
    const pasted = e.clipboardData?.files
    if (pasted && pasted.length > 0) addFiles(Array.from(pasted))
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault()
    setDragOver(false)
    const dropped = e.dataTransfer?.files
    if (dropped && dropped.length > 0) addFiles(Array.from(dropped))
  }

  function handleDragOver(e: React.DragEvent<HTMLDivElement>) {
    if (editingMessageId) return
    e.preventDefault()
    setDragOver(true)
  }

  function handleDragLeave(e: React.DragEvent<HTMLDivElement>) {
    // Only clear when the pointer actually leaves the composer, not a child.
    if (e.currentTarget.contains(e.relatedTarget as Node | null)) return
    setDragOver(false)
  }

  function submit() {
    const content = value.trim()
    if (editingMessageId) {
      if (!content) return
      submitEdit(editingMessageId, content)
    } else if (files.length > 0) {
      // Attachments allow an empty caption.
      sendWithAttachments({
        content,
        files,
        replyToMessageId: replyTo ?? undefined,
        mentions: resolveMentions(content, members),
      })
      setReplyTo(null)
      setFiles([])
    } else if (content) {
      send({
        content,
        replyToMessageId: replyTo ?? undefined,
        mentions: resolveMentions(content, members),
      })
      setReplyTo(null)
    } else {
      return
    }
    setValue("")
    setMentionQuery(null)
    caretPosRef.current = 0
    pendingCaretRef.current = 0
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    // Mention popover captures navigation + selection keys first.
    if (mentionOpen) {
      if (e.key === "ArrowDown") {
        e.preventDefault()
        setMentionIndex((i) => (i + 1) % mentionMatches.length)
        return
      }
      if (e.key === "ArrowUp") {
        e.preventDefault()
        setMentionIndex((i) => (i - 1 + mentionMatches.length) % mentionMatches.length)
        return
      }
      if (e.key === "Enter" || e.key === "Tab") {
        e.preventDefault()
        commitMention(mentionMatches[mentionIndex])
        return
      }
      if (e.key === "Escape") {
        e.preventDefault()
        setMentionQuery(null)
        return
      }
    }

    // Escape cancels an in-progress edit (when no popover claimed it first).
    if (e.key === "Escape" && editingMessageId && !mentionOpen && !emojiOpen) {
      e.preventDefault()
      cancelEdit()
      return
    }

    // Enter-guard: never send while a popover (mention OR emoji) is open.
    if (e.key === "Enter" && !e.shiftKey && !mentionOpen && !emojiOpen) {
      e.preventDefault()
      submit()
    }
  }

  const canAttach = !editingMessageId && files.length < MAX_MEDIA_PER_MESSAGE
  const canSend = value.trim().length > 0 || (files.length > 0 && !editingMessageId)

  return (
    <div className="border-t px-3 py-3">
      <div
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        className={cn(
          "bg-card focus-within:ring-ring/40 relative flex flex-col gap-1.5 rounded-2xl border p-1.5 shadow-sm transition focus-within:ring-2",
          dragOver && "ring-primary ring-2"
        )}
      >
        {/* Editing strip — supersedes the reply strip; X (or Esc) cancels. */}
        {editingMessage && (
          <div className="border-primary bg-muted/50 flex items-center gap-2 rounded-lg border-s-2 px-2.5 py-1.5">
            <PencilIcon className="text-primary size-3.5 shrink-0" />
            <div className="flex min-w-0 flex-col">
              <span className="text-primary text-xs font-semibold">{t("conversations.composer.editing")}</span>
              <span className="text-muted-foreground truncate text-xs">{editingMessage.content}</span>
            </div>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="ms-auto shrink-0"
              aria-label={t("conversations.composer.cancelEdit")}
              onClick={cancelEdit}
            >
              <XIcon />
            </Button>
          </div>
        )}

        {/* Reply strip — accent start-border, sender + snippet, X to clear. */}
        {!editingMessage && replyMessage && (
          <div className="border-primary bg-muted/50 flex items-center gap-2 rounded-lg border-s-2 px-2.5 py-1.5">
            <div className="flex min-w-0 flex-col">
              <span className="text-primary text-xs font-semibold">
                {t("conversations.composer.replyingTo", { name: replyMessage.sender?.name ?? "" })}
              </span>
              <span className="text-muted-foreground truncate text-xs">{replyMessage.content}</span>
            </div>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              className="ms-auto shrink-0"
              aria-label={t("conversations.composer.cancelReply")}
              onClick={() => setReplyTo(null)}
            >
              <XIcon />
            </Button>
          </div>
        )}

        {mentionOpen && (
          <MentionPopover
            members={mentionMatches}
            activeIndex={mentionIndex}
            onSelect={commitMention}
            onHover={setMentionIndex}
          />
        )}

        {/* Pre-send attachment tray — hidden while editing (no media edits). */}
        {!editingMessageId && files.length > 0 && <AttachmentTray files={files} onRemove={removeFile} />}

        <input
          ref={fileInputRef}
          type="file"
          multiple
          className="hidden"
          onChange={onFilePick}
          aria-hidden
          tabIndex={-1}
        />

        <div className="flex items-end gap-1">
          {/* Attach files — disabled while editing or at the media cap. */}
          <Button
            type="button"
            variant="ghost"
            size="icon-sm"
            className="text-muted-foreground shrink-0"
            disabled={!canAttach}
            aria-label={t("conversations.composer.attach")}
            onClick={() => fileInputRef.current?.click()}
          >
            <PaperclipIcon />
          </Button>

          <textarea
            ref={textareaRef}
            value={value}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            onKeyUp={handleCaretSync}
            onClick={handleCaretSync}
            onPaste={handlePaste}
            rows={1}
            placeholder={t("conversations.composer.placeholder")}
            aria-label={t("conversations.composer.placeholder")}
            className="text-foreground placeholder:text-muted-foreground max-h-40 min-h-9 flex-1 resize-none bg-transparent px-2 py-1.5 text-sm leading-relaxed outline-none"
          />

          <Popover open={emojiOpen} onOpenChange={setEmojiOpen}>
            <PopoverTrigger
              render={
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  className="text-muted-foreground shrink-0"
                  aria-label={t("conversations.composer.emoji")}
                />
              }
            >
              <SmileIcon />
            </PopoverTrigger>
            <PopoverContent align="end" side="top" className="w-fit p-0">
              <EmojiPicker.Root
                onEmojiSelect={({ emoji }) => insertEmoji(emoji)}
                className="isolate flex h-80 w-72 flex-col"
              >
                <EmojiPicker.Search
                  placeholder={t("conversations.composer.emojiSearch")}
                  className="bg-muted/60 placeholder:text-muted-foreground focus-visible:ring-ring/40 m-2 rounded-lg px-2.5 py-2 text-sm outline-none focus-visible:ring-2"
                />
                <EmojiPicker.Viewport className="relative flex-1 outline-hidden">
                  <EmojiPicker.Loading className="text-muted-foreground absolute inset-0 flex items-center justify-center text-sm">
                    {t("conversations.composer.emojiLoading")}
                  </EmojiPicker.Loading>
                  <EmojiPicker.Empty className="text-muted-foreground absolute inset-0 flex items-center justify-center text-sm">
                    {t("conversations.composer.emojiEmpty")}
                  </EmojiPicker.Empty>
                  <EmojiPicker.List
                    className="pb-2 select-none"
                    components={{
                      CategoryHeader: ({ category, ...props }) => (
                        <div className="bg-popover text-muted-foreground px-2 pt-2 pb-1 text-xs font-medium" {...props}>
                          {category.label}
                        </div>
                      ),
                      Row: ({ children, ...props }) => (
                        <div className="scroll-my-1 px-1" {...props}>
                          {children}
                        </div>
                      ),
                      Emoji: ({ emoji, ...props }) => (
                        <button
                          className={cn(
                            "flex size-8 items-center justify-center rounded-md text-lg",
                            emoji.isActive && "bg-accent"
                          )}
                          {...props}
                        >
                          {emoji.emoji}
                        </button>
                      ),
                    }}
                  />
                </EmojiPicker.Viewport>
              </EmojiPicker.Root>
            </PopoverContent>
          </Popover>

          <Button
            type="button"
            size="icon-sm"
            className="shrink-0"
            disabled={!canSend}
            aria-label={editingMessageId ? t("conversations.composer.saveEdit") : t("conversations.composer.send")}
            onClick={submit}
          >
            {editingMessageId ? <CheckIcon /> : <SendHorizontalIcon className="rtl:rotate-180" />}
          </Button>
        </div>
      </div>
    </div>
  )
}
