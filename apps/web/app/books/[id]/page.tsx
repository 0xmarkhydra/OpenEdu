import Link from "next/link";
import { notFound } from "next/navigation";

type Book = { id: string; title: string; grade: number; subject_code: string; rights_owner: string; rights_license: string; rights_source: string; status: string };

export default async function BookPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  if (!/^[a-fA-F0-9-]{36}$/.test(id)) notFound();
  let book: Book;
  try {
    const res = await fetch((process.env.API_INTERNAL_URL || "http://localhost:8080") + "/v1/books/" + encodeURIComponent(id), {
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    });
    if (!res.ok) notFound();
    book = await res.json() as Book;
  } catch {
    notFound();
  }
  return (
    <section className="container detail">
      <Link href="/" className="back">← Trở về thư viện</Link>
      <div className="detail-card">
        <div className="detail-cover" aria-hidden="true">📖</div>
        <div>
          <span className="eyebrow section-eyebrow">HỌC LIỆU ĐÃ PHÁT HÀNH · LỚP {book.grade}</span>
          <h1>{book.title}</h1>
          <p className="detail-subject">Môn: {book.subject_code}</p>
          <dl>
            <dt>Đơn vị sở hữu</dt><dd>{book.rights_owner}</dd>
            <dt>Quyền sử dụng</dt><dd>{book.rights_license}</dd>
            <dt>Nguồn</dt><dd>{book.rights_source}</dd>
          </dl>
          <p className="detail-notice">Nội dung chương, bài và trình đọc PDF sẽ được mở khi hệ thống quản trị học liệu và quyền phân phối hoàn tất.</p>
        </div>
      </div>
    </section>
  );
}
