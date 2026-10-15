import type { Metadata } from "next";
import Link from "next/link";
import { BookOpenCheck, LibraryBig, LayoutDashboard, LogIn } from "lucide-react";
import "./globals.css";

export const metadata: Metadata = {
  title: "OpenEdu — Học tập và kết nối",
  description: "Thư viện học liệu số và nền tảng kết nối nhà trường, giáo viên, học sinh và phụ huynh.",
};

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="vi">
      <body>
        <header className="site-header">
          <div className="container header-inner">
            <Link className="logo" href="/" aria-label="OpenEdu - về trang chủ">
              <span className="logo-icon" aria-hidden="true"><BookOpenCheck size={25} strokeWidth={1.9}/></span>
              <span>Open<span className="logo-accent">Edu</span></span>
            </Link>
            <nav aria-label="Điều hướng chính">
              <Link href="/"><LibraryBig size={17} strokeWidth={1.9} aria-hidden="true"/>Thư viện</Link>
              <Link href="/dashboard"><LayoutDashboard size={17} strokeWidth={1.9} aria-hidden="true"/>Không gian học tập</Link>
              <Link href="/login"><LogIn size={17} strokeWidth={1.9} aria-hidden="true"/>Đăng nhập</Link>
            </nav>
          </div>
        </header>
        <main>{children}</main>
        <footer className="site-footer container">
          <span>© 2026 OpenEdu Core</span>
          <span>Học liệu được hiển thị theo quyền sử dụng và bản quyền.</span>
        </footer>
      </body>
    </html>
  );
}
