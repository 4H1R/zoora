import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { DownloadIcon, ExternalLinkIcon, Loader2Icon } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { toast } from "sonner"

import {
  getAdminTutorialsAparatOembed,
  useGetAdminTutorialsId,
  usePostAdminTutorialsIdPublish,
  usePostAdminTutorialsIdUnpublish,
  usePutAdminTutorialsId,
} from "@/api/admin-tutorials/admin-tutorials"
import { ChangelogMarkdown } from "@/components/changelog/markdown"
import { aparatEmbedUrl, aparatWatchUrl, extractAparatHash } from "@/components/tutorials/aparat"
import { PageHeader } from "@/components/page-header"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Textarea } from "@/components/ui/textarea"

export const Route = createFileRoute("/_admin/admin/tutorials/$id")({
  component: TutorialEditorPage,
})

function TutorialEditorPage() {
  const { id } = Route.useParams()
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const { data } = useGetAdminTutorialsId(id)
  const tu = (data?.status === 200 && data.data.data) || undefined

  const [form, setForm] = useState({
    title_en: "",
    title_fa: "",
    description_en: "",
    description_fa: "",
    aparat_hash: "",
    thumbnail_url: "",
  })
  const [linkInput, setLinkInput] = useState("")
  const [fetching, setFetching] = useState(false)

  useEffect(() => {
    if (tu) {
      setForm({
        title_en: tu.title_en ?? "",
        title_fa: tu.title_fa ?? "",
        description_en: tu.description_en ?? "",
        description_fa: tu.description_fa ?? "",
        aparat_hash: tu.aparat_hash && tu.aparat_hash !== "placeholder" ? tu.aparat_hash : "",
        thumbnail_url: tu.thumbnail_url ?? "",
      })
      if (tu.aparat_hash && tu.aparat_hash !== "placeholder") setLinkInput(aparatWatchUrl(tu.aparat_hash))
    }
  }, [tu])

  const save = usePutAdminTutorialsId({
    mutation: { onSuccess: () => toast.success(t("admin.tutorials.saved")) },
  })
  const publish = usePostAdminTutorialsIdPublish({
    mutation: {
      onSuccess: () => {
        toast.success(t("admin.tutorials.publishedToast"))
        queryClient.invalidateQueries()
      },
    },
  })
  const unpublish = usePostAdminTutorialsIdUnpublish({
    mutation: { onSuccess: () => queryClient.invalidateQueries() },
  })

  // Paste an Aparat link → pull the hash → ask Aparat for the poster + title.
  async function loadFromAparat() {
    const hash = extractAparatHash(linkInput)
    if (!hash) {
      toast.error(t("admin.tutorials.badLink"))
      return
    }
    setForm((f) => ({ ...f, aparat_hash: hash }))
    setFetching(true)
    // Proxied through our backend — Aparat's oEmbed sends no CORS headers, so
    // the browser can't read it directly.
    const res = await getAdminTutorialsAparatOembed({ hash }).catch(() => null)
    setFetching(false)
    const meta = res?.status === 200 ? res.data.data : undefined
    if (!meta) {
      toast.warning(t("admin.tutorials.oembedFailed"))
      return
    }
    setForm((f) => ({
      ...f,
      thumbnail_url: meta.thumbnail_url ?? f.thumbnail_url,
      // Only suggest a title when the field is still empty — never clobber.
      title_en: f.title_en || meta.title || f.title_en,
    }))
    toast.success(t("admin.tutorials.oembedOk"))
  }

  function doSave() {
    if (!form.aparat_hash) {
      toast.error(t("admin.tutorials.needVideo"))
      return
    }
    save.mutate({
      id,
      data: {
        title_en: form.title_en,
        title_fa: form.title_fa,
        description_en: form.description_en,
        description_fa: form.description_fa,
        aparat_hash: form.aparat_hash,
        thumbnail_url: form.thumbnail_url,
      },
    })
  }

  const isPublished = !!tu?.published_at
  const canPublish = !!form.aparat_hash

  return (
    <div className="mx-auto w-full max-w-5xl space-y-5">
      <PageHeader
        title={t("admin.tutorials.editTitle")}
        actions={
          <div className="flex gap-2">
            <Button variant="outline" onClick={doSave} disabled={save.isPending}>
              {t("admin.tutorials.saveDraft")}
            </Button>
            {isPublished ? (
              <Button variant="secondary" onClick={() => unpublish.mutate({ id })}>
                {t("admin.tutorials.unpublish")}
              </Button>
            ) : (
              <Button onClick={() => publish.mutate({ id })} disabled={!canPublish}>
                {t("admin.tutorials.publish")}
              </Button>
            )}
          </div>
        }
      />

      {/* Video source */}
      <div className="space-y-3 rounded-xl border p-4">
        <Label>{t("admin.tutorials.aparatLink")}</Label>
        <div className="flex gap-2">
          <Input
            value={linkInput}
            placeholder="https://www.aparat.com/v/AbCdE"
            onChange={(e) => setLinkInput(e.target.value)}
            dir="ltr"
          />
          <Button variant="outline" onClick={loadFromAparat} disabled={fetching || !linkInput.trim()}>
            {fetching ? (
              <Loader2Icon className="size-4 animate-spin" />
            ) : (
              <DownloadIcon data-icon="inline-start" />
            )}
            {t("admin.tutorials.load")}
          </Button>
        </div>
        <p className="text-muted-foreground text-xs">{t("admin.tutorials.aparatHint")}</p>

        {form.aparat_hash && (
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5">
              <span className="text-muted-foreground text-xs font-medium">{t("admin.tutorials.preview")}</span>
              <div className="bg-muted overflow-hidden rounded-lg border">
                <iframe
                  key={form.aparat_hash}
                  src={aparatEmbedUrl(form.aparat_hash)}
                  title={t("admin.tutorials.preview")}
                  allow="fullscreen"
                  className="aspect-video w-full border-0"
                />
              </div>
              <a
                href={aparatWatchUrl(form.aparat_hash)}
                target="_blank"
                rel="noreferrer"
                className="text-muted-foreground hover:text-foreground inline-flex items-center gap-1 text-xs"
              >
                <ExternalLinkIcon className="size-3" />
                {t("admin.tutorials.openOnAparat")}
              </a>
            </div>

            <div className="space-y-1.5">
              <Label className="text-xs">{t("admin.tutorials.thumbnail")}</Label>
              <Input
                value={form.thumbnail_url}
                placeholder="https://…/poster.jpg"
                onChange={(e) => setForm((f) => ({ ...f, thumbnail_url: e.target.value }))}
                dir="ltr"
              />
              {form.thumbnail_url && (
                <img
                  src={form.thumbnail_url}
                  alt=""
                  className="aspect-video w-full rounded-lg border object-cover"
                />
              )}
            </div>
          </div>
        )}
      </div>

      {/* Localized text */}
      <Tabs defaultValue="en">
        <TabsList>
          <TabsTrigger value="en">EN</TabsTrigger>
          <TabsTrigger value="fa">FA</TabsTrigger>
        </TabsList>

        <TabsContent value="en" className="space-y-3">
          <div className="space-y-1">
            <Label>{t("admin.tutorials.titleLabel")}</Label>
            <Input
              value={form.title_en}
              placeholder={t("admin.tutorials.titlePh")}
              onChange={(e) => setForm((f) => ({ ...f, title_en: e.target.value }))}
            />
          </div>
          <div className="space-y-1">
            <Label>{t("admin.tutorials.descriptionLabel")}</Label>
            <div className="grid grid-cols-2 gap-3">
              <Textarea
                className="min-h-[240px] font-mono text-sm"
                value={form.description_en}
                placeholder={t("admin.tutorials.descriptionPh")}
                onChange={(e) => setForm((f) => ({ ...f, description_en: e.target.value }))}
              />
              <div className="rounded-md border p-3">
                <ChangelogMarkdown>{form.description_en}</ChangelogMarkdown>
              </div>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="fa" className="space-y-3" dir="rtl">
          <div className="space-y-1">
            <Label>{t("admin.tutorials.titleLabel")}</Label>
            <Input
              value={form.title_fa}
              placeholder={t("admin.tutorials.titlePh")}
              onChange={(e) => setForm((f) => ({ ...f, title_fa: e.target.value }))}
            />
          </div>
          <div className="space-y-1">
            <Label>{t("admin.tutorials.descriptionLabel")}</Label>
            <div className="grid grid-cols-2 gap-3">
              <Textarea
                className="min-h-[240px] font-mono text-sm"
                value={form.description_fa}
                placeholder={t("admin.tutorials.descriptionPh")}
                onChange={(e) => setForm((f) => ({ ...f, description_fa: e.target.value }))}
              />
              <div className="rounded-md border p-3" dir="rtl">
                <ChangelogMarkdown>{form.description_fa}</ChangelogMarkdown>
              </div>
            </div>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
