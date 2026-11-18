import { Dialog } from "./Dialog";
import { CategoryIcon } from "./FileIcon";
import { useI18n } from "../i18n";
import { formatBytes } from "../lib/format";
import { Download } from "../lib/types";

interface Props {
  d: Download;
  onOpenFile: () => void;
  onOpenFolder: () => void;
  onClose: () => void;
  /** Fill the window without an overlay (the popup). */
  bare?: boolean;
}

/** "Download complete" (ADR 0023): the same look as the quick-add popup. */
export function DoneDialog({ d, onOpenFile, onOpenFolder, onClose, bare }: Props) {
  const { t } = useI18n();
  const size = d.size >= 0 ? d.size : d.downloaded;
  return (
    <Dialog
      title={`✅ ${t("toast.completed")}`}
      onClose={onClose}
      bare={bare}
      testId="done-dialog"
      footer={
        <>
          <button className="btn ghost" onClick={onClose} data-testid="done-close">
            {t("done.close")}
          </button>
          <button className="btn" onClick={onOpenFolder} data-testid="done-folder">
            <span className="emoji" aria-hidden="true">📂</span> {t("done.openFolder")}
          </button>
          {!d.isDir && (
            <button className="btn primary" onClick={onOpenFile} data-testid="done-open">
              <span className="emoji" aria-hidden="true">📄</span> {t("done.openFile")}
            </button>
          )}
        </>
      }
    >
      <div className="form">
        <div className="quick-file" data-testid="done-file">
          <CategoryIcon category={d.category} size={30} />
          <div className="probe-info">
            <strong className="ellipsis" title={d.filename}>
              <bdi>{d.filename}</bdi>
            </strong>
            <span className="muted">
              <bdi>{size > 0 ? formatBytes(size) : t("add.unknownSize")}</bdi>
            </span>
          </div>
        </div>
        <p className="muted ellipsis done-dir" title={d.dir}>
          <bdi>{d.dir}</bdi>
        </p>
      </div>
    </Dialog>
  );
}
