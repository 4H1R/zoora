import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  useGetAdminChangelogId,
  usePostAdminChangelogIdPublish,
  usePostAdminChangelogIdUnpublish,
  usePostAdminChangelogMediaPresign,
  usePutAdminChangelogId,
} from "@/api/admin-changelog/admin-changelog"
import { ChangelogMarkdown } from "@/components/changelog/markdown"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"

export const Route = createFileRoute("/_admin/admin/changelog/$id")({
  component: ChangelogEditorPage,
})

function ChangelogEditorPage() {
  const { id } = Route.useParams()
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { data } = useGetAdminChangelogId(id)
  const entry = (data?.status === 200 && data.data.data) || undefined

  const [form, setForm] = useState({
    version: "",
    title_en: "",
    title_fa: "",
    body_en: "",
    body_fa: "",
    is_major: false,
  })

  useEffect(() => {
    if (entry)
      setForm({
        version: entry.version ?? "",
        title_en: entry.title_en ?? "",
        title_fa: entry.title_fa ?? "",
        body_en: entry.body_en ?? "",
        body_fa: entry.body_fa ?? "",
        is_major: entry.is_major ?? false,
      })
  }, [entry])

  const save = usePutAdminChangelogId({
    mutation: { onSuccess: () => toast.success(t("admin.changelog.saved")) },
  })
  const publish = usePostAdminChangelogIdPublish({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.changelog.publishedToast"))
        queryClient.invalidateQueries()
      },
    },
  })
  const unpublish = usePostAdminChangelogIdUnpublish({
    mutation: { onSuccess: () => queryClient.invalidateQueries() },
  })
  const presign = usePostAdminChangelogMediaPresign()

  function doSave() {
    save.mutate({
      id,
      data: {
        version: form.version || undefined,
        title_en: form.title_en,
        title_fa: form.title_fa,
        body_en: form.body_en,
        body_fa: form.body_fa,
        is_major: form.is_major,
      },
    })
  }

  // Presign → PUT the file straight to S3 → insert markdown token at cursor.
  async function uploadMedia(file: File, target: "en" | "fa") {
    const res = await presign.mutateAsync({
      data: { entry_id: id, file_name: file.name, mime_type: file.type, size: file.size },
    })
    if (res.status !== 200) {
      toast.error(t("admin.changelog.uploadFailed"))
      return
    }
    const { upload_url, public_url } = res.data.data ?? {}
    if (!upload_url || !public_url) {
      toast.error(t("admin.changelog.uploadFailed"))
      return
    }
    const put = await fetch(upload_url, {
      method: "PUT",
      headers: { "Content-Type": file.type },
      body: file,
    })
    if (!put.ok) {
      toast.error(t("admin.changelog.uploadFailed"))
      return
    }
    const token = `\n\n![${file.name}](${public_url})\n\n`
    const key = target === "en" ? "body_en" : "body_fa"
    setForm((f) => ({ ...f, [key]: f[key] + token }))
  }

  const isPublished = !!entry?.published_at

  return (
    <div className="mx-auto w-full max-w-5xl space-y-4">
      <PageHeader
        title={t("admin.changelog.editTitle")}
        actions={
          <div className="flex gap-2">
            <Button variant="outline" onClick={doSave}>
              {t("admin.changelog.saveDraft")}
            </Button>
            {isPublished ? (
              <Button variant="secondary" onClick={() => unpublish.mutate({ id })}>
                {t("admin.changelog.unpublish")}
              </Button>
            ) : (
              <Button onClick={() => publish.mutate({ id })}>
                {t("admin.changelog.publish")}
              </Button>
            )}
          </div>
        }
      />

      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-1">
          <Label>{t("admin.changelog.version")}</Label>
          <Input
            value={form.version}
            placeholder="v2.4.0"
            onChange={(e) => setForm((f) => ({ ...f, version: e.target.value }))}
          />
        </div>
        <label className="flex items-end gap-2 pb-2">
          <Checkbox
            checked={form.is_major}
            onCheckedChange={(v) => setForm((f) => ({ ...f, is_major: !!v }))}
          />
          {t("admin.changelog.major")}
        </label>
      </div>

      <Tabs defaultValue="en">
        <TabsList>
          <TabsTrigger value="en">EN</TabsTrigger>
          <TabsTrigger value="fa">FA</TabsTrigger>
        </TabsList>

        <TabsContent value="en" className="space-y-3">
          <Input
            value={form.title_en}
            placeholder={t("admin.changelog.titlePh")}
            onChange={(e) => setForm((f) => ({ ...f, title_en: e.target.value }))}
          />
          <MediaUpload onPick={(file) => uploadMedia(file, "en")} label={t("admin.changelog.insertMedia")} />
          <div className="grid grid-cols-2 gap-3">
            <Textarea
              className="min-h-[400px] font-mono text-sm"
              value={form.body_en}
              onChange={(e) => setForm((f) => ({ ...f, body_en: e.target.value }))}
            />
            <div className="rounded-md border p-3">
              <ChangelogMarkdown>{form.body_en}</ChangelogMarkdown>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="fa" className="space-y-3" dir="rtl">
          <Input
            value={form.title_fa}
            placeholder={t("admin.changelog.titlePh")}
            onChange={(e) => setForm((f) => ({ ...f, title_fa: e.target.value }))}
          />
          <MediaUpload onPick={(file) => uploadMedia(file, "fa")} label={t("admin.changelog.insertMedia")} />
          <div className="grid grid-cols-2 gap-3">
            <Textarea
              className="min-h-[400px] font-mono text-sm"
              value={form.body_fa}
              onChange={(e) => setForm((f) => ({ ...f, body_fa: e.target.value }))}
            />
            <div className="rounded-md border p-3" dir="rtl">
              <ChangelogMarkdown>{form.body_fa}</ChangelogMarkdown>
            </div>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}

function MediaUpload({ onPick, label }: { onPick: (f: File) => void; label: string }) {
  const [ref, setRef] = useState<HTMLInputElement | null>(null)
  return (
    <>
      <Button variant="outline" size="sm" onClick={() => ref?.click()}>
        {label}
      </Button>
      <input
        ref={setRef}
        type="file"
        accept="image/*,video/*"
        className="hidden"
        onChange={(e) => {
          const f = e.target.files?.[0]
          if (f) onPick(f)
          e.target.value = ""
        }}
      />
    </>
  )
}
