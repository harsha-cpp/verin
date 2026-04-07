import { Upload } from "lucide-react";
import { useEffect, useRef, useState } from "react";

export function DragUploadOverlay({ onFileDrop }: { onFileDrop: (file: File) => void }) {
  const [visible, setVisible] = useState(false);
  const dragDepth = useRef(0);

  useEffect(() => {
    function onDragEnter(e: DragEvent) {
      if (!e.dataTransfer?.types.includes("Files")) return;
      e.preventDefault();
      dragDepth.current += 1;
      if (dragDepth.current === 1) setVisible(true);
    }

    function onDragLeave() {
      dragDepth.current -= 1;
      if (dragDepth.current === 0) setVisible(false);
    }

    function onDragOver(e: DragEvent) {
      if (e.dataTransfer?.types.includes("Files")) e.preventDefault();
    }

    function onDrop(e: DragEvent) {
      e.preventDefault();
      dragDepth.current = 0;
      setVisible(false);
      const file = e.dataTransfer?.files[0];
      if (file) onFileDrop(file);
    }

    window.addEventListener("dragenter", onDragEnter);
    window.addEventListener("dragleave", onDragLeave);
    window.addEventListener("dragover", onDragOver);
    window.addEventListener("drop", onDrop);

    return () => {
      window.removeEventListener("dragenter", onDragEnter);
      window.removeEventListener("dragleave", onDragLeave);
      window.removeEventListener("dragover", onDragOver);
      window.removeEventListener("drop", onDrop);
    };
  }, [onFileDrop]);

  if (!visible) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      style={{
        background: "rgba(250,248,245,0.88)",
        backdropFilter: "blur(8px)",
        animation: "fade-in 0.15s ease both",
      }}
    >
      <div
        className="flex flex-col items-center gap-5 rounded-3xl px-16 py-14"
        style={{
          border: "2px dashed var(--accent)",
          background: "var(--accent-soft)",
        }}
      >
        <div
          className="flex h-16 w-16 items-center justify-center rounded-2xl processing-pulse"
          style={{ background: "var(--accent-soft)", border: "2px solid var(--accent)" }}
        >
          <Upload className="h-7 w-7" style={{ color: "var(--accent-strong)" }} />
        </div>
        <div className="text-center">
          <div className="font-display text-2xl" style={{ color: "var(--ink)" }}>
            Drop to upload
          </div>
          <p className="mt-2 text-sm" style={{ color: "var(--ink-soft)" }}>
            Release to add this file to your workspace
          </p>
        </div>
      </div>
    </div>
  );
}
