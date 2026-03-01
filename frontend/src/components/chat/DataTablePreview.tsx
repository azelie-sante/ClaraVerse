/**
 * DataTablePreview Component
 *
 * Renders a preview of CSV/tabular data as an inline table widget.
 * Shows headers, first few rows, and metadata (row count).
 */

import { Download, Maximize2 } from 'lucide-react';
import type { DataPreview } from '@/types/websocket';
import { Tooltip } from '@/components/design-system/Tooltip/Tooltip';
import styles from './DataTablePreview.module.css';

interface DataTablePreviewProps {
  filename: string;
  preview: DataPreview;
  onDownload?: () => void;
  onExpand?: () => void;
}

export const DataTablePreview = ({
  filename,
  preview,
  onDownload,
  onExpand,
}: DataTablePreviewProps) => {
  // Extract title from filename (remove extension)
  const title = filename.replace(/\.[^/.]+$/, '').replace(/_/g, ' ');

  return (
    <div className={styles.container}>
      {/* Header */}
      <div className={styles.header}>
        <h3 className={styles.title}>{title}</h3>
        <div className={styles.actions}>
          {onDownload && (
            <Tooltip content="Download file" position="top">
              <button onClick={onDownload} className={styles.actionButton}>
                <Download size={16} />
              </button>
            </Tooltip>
          )}
          {onExpand && (
            <Tooltip content="Expand view" position="top">
              <button onClick={onExpand} className={styles.actionButton}>
                <Maximize2 size={16} />
              </button>
            </Tooltip>
          )}
        </div>
      </div>

      {/* Table */}
      <div className={styles.tableWrapper}>
        <table className={styles.table}>
          <thead>
            <tr>
              <th className={styles.rowNumber}></th>
              {preview.headers.map((header, idx) => (
                <th key={idx} className={styles.headerCell}>
                  {header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {preview.rows.map((row, rowIdx) => (
              <tr key={rowIdx} className={styles.row}>
                <td className={styles.rowNumber}>{rowIdx + 1}</td>
                {row.map((cell, cellIdx) => (
                  <td key={cellIdx} className={styles.cell}>
                    {cell}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Footer with metadata */}
      {preview.row_count > preview.rows.length && (
        <div className={styles.footer}>
          Showing {preview.rows.length} of {preview.row_count} rows
        </div>
      )}
    </div>
  );
};
