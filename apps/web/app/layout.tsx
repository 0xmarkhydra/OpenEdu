import type { Metadata } from "next";
import Link from "next/link";
import "./globals.css";

export const metadata: Metadata = {
  title: "FreeEducation — Học tập và kết nối",
  description: "Thư viện học liệu số và nền tảng kết nối nhà trường, giáo viên, học sinh và phụ huynh.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="vi">
      <body>
        <header className="site-header">
          <div className="container header-inner">
            <Link className="logo" href="/" aria-label="FreeEducation - về trang chủ">
              <span className="logo-icon" aria-hidden="true">F</span>
              <span>Free<span className="logo-accent">Education</span></span>
            </Link>
            <nav aria-label="Điều hướng chính">
              <Link href="/">Thư viện</Link>
              <span className="nav-disabled" title="Đang phát triển">Học tập</span>
              <span className="nav-disabled" title="Đang phát triển">Sổ liên lạc</span>
            </nav>
          </div>
        </header>
        <main>{children}</main>
        <footer className="site-footer container">
          <span>© 2026 FreeEducation Core</span>
          <span>Học liệu được hiển thị theo quyền sử dụng và bản quyền.</span>
        </footer>
      </body>
    </html>
  );
}
