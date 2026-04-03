import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  type SortingState,
  useReactTable,
} from "@tanstack/react-table";
import { ArrowUpDown, FileText } from "lucide-react";
import { useState } from "react";
import { Badge, Table, TableBody, TableCell, TableHead, TableHeaderCell, TableRow } from "@verin/ui";
import { Link } from "react-router-dom";

import { formatBytes, formatDate, getStatusTone } from "@/lib/utils";

type DocumentRow = {
  id: string;
  title: string;
  originalFilename: string;
  status: string;
  currentVersionNumber?: number;
  sizeBytes?: number;
  updatedAt?: string;
};

const columnHelper = createColumnHelper<DocumentRow>();

const columns = [
  columnHelper.accessor("title", {
    header: "Document",
    cell: ({ row, getValue }) => (
      <Link to={`/documents/${row.original.id}`} className="flex items-center gap-3">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-slate-100 text-slate-500">
          <FileText className="h-3.5 w-3.5" />
        </div>
        <div className="min-w-0">
          <div className="truncate text-sm font-medium text-slate-900">{getValue()}</div>
          <div className="truncate text-xs text-slate-500">{row.original.originalFilename}</div>
        </div>
      </Link>
    ),
    enableSorting: true,
  }),
  columnHelper.accessor("status", {
    header: "Status",
    cell: ({ getValue }) => <Badge tone={getStatusTone(getValue())}>{getValue()}</Badge>,
    enableSorting: true,
  }),
  columnHelper.accessor("currentVersionNumber", {
    header: "Version",
    cell: ({ getValue }) => <span className="text-sm text-slate-600">v{getValue() ?? 1}</span>,
    enableSorting: true,
  }),
  columnHelper.accessor("sizeBytes", {
    header: "Size",
    cell: ({ getValue }) => <span className="text-sm text-slate-500">{formatBytes(getValue() ?? 0)}</span>,
    enableSorting: true,
  }),
  columnHelper.accessor("updatedAt", {
    header: "Updated",
    cell: ({ getValue }) => <span className="text-sm text-slate-500">{formatDate(getValue())}</span>,
    enableSorting: true,
  }),
];

export function DocumentTable({ data }: { data: DocumentRow[] }) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const table = useReactTable({
    data,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });

  if (!data.length) {
    return (
      <div className="flex min-h-[200px] flex-col items-center justify-center rounded-xl border border-slate-200 bg-white py-10 text-center">
        <div className="flex h-10 w-10 items-center justify-center rounded-md bg-slate-100 text-slate-400">
          <FileText className="h-5 w-5" />
        </div>
        <h3 className="mt-3 text-sm font-semibold text-slate-900">No documents yet</h3>
        <p className="mt-1 max-w-xs text-sm text-slate-500">
          Upload documents to create a searchable, versioned workspace.
        </p>
      </div>
    );
  }

  return (
    <div className="overflow-hidden rounded-xl border border-slate-200 bg-white">
      <Table>
        <TableHead>
          {table.getHeaderGroups().map((group) => (
            <TableRow key={group.id}>
              {group.headers.map((header) => (
                <TableHeaderCell
                  key={header.id}
                  className={header.column.getCanSort() ? "cursor-pointer select-none" : ""}
                  onClick={header.column.getToggleSortingHandler()}
                >
                  <span className="inline-flex items-center gap-1">
                    {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                    {header.column.getCanSort() && <ArrowUpDown className="h-3 w-3 text-slate-300" />}
                  </span>
                </TableHeaderCell>
              ))}
            </TableRow>
          ))}
        </TableHead>
        <TableBody>
          {table.getRowModel().rows.map((row) => (
            <TableRow key={row.id} className="transition-colors hover:bg-slate-50">
              {row.getVisibleCells().map((cell) => (
                <TableCell key={cell.id}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
