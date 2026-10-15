"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, ArrowRight, BookOpenCheck } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function LoginPage() {
  const router=useRouter();
  const [email,setEmail]=useState("");
  const [password,setPassword]=useState("");
  const [school,setSchool]=useState("");
  const [busy,setBusy]=useState(false);
  const [error,setError]=useState("");
  async function submit(event:FormEvent<HTMLFormElement>){
    event.preventDefault();
    setError("");setBusy(true);
    try{
      const response=await fetch("/api/auth/login",{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({email,password,school_id:school})});
      const data=await response.json() as {error?:string};
      if(!response.ok) {setError(data.error||"Không thể đăng nhập");return;}
      setPassword("");router.replace("/dashboard");router.refresh();
    }catch{setError("Không thể kết nối máy chủ. Vui lòng thử lại.");}
    finally{setBusy(false);}
  }
  return <section className="auth-wrap">
    <div className="auth-card">
      <Link className="back" href="/"><ArrowLeft size={16} aria-hidden="true"/>Thư viện OpenEdu</Link>
      <div className="auth-symbol" aria-hidden="true"><BookOpenCheck size={28} strokeWidth={1.9}/></div>
      <span className="eyebrow section-eyebrow">OPENEDU ACCOUNT</span>
      <h1>Đăng nhập</h1>
      <p>Không gian dành cho quản trị nhà trường và các tài khoản đã được cấp quyền.</p>
      <form onSubmit={submit} className="auth-form">
        <label>Email<input type="email" autoComplete="username" required maxLength={254} value={email} onChange={e=>setEmail(e.target.value)} placeholder="email@truong.edu.vn"/></label>
        <label>Mật khẩu<input type="password" autoComplete="current-password" minLength={12} required value={password} onChange={e=>setPassword(e.target.value)} placeholder="Nhập mật khẩu"/></label>
        <label>Mã trường <small>(chỉ cần nếu tài khoản thuộc nhiều trường)</small><input value={school} onChange={e=>setSchool(e.target.value.trim())} placeholder="UUID của trường (không bắt buộc)"/></label>
        {error && <div role="alert" className="form-error">{error==="select a school to sign in"?"Vui lòng chọn mã trường trước khi đăng nhập.":error}</div>}
        <Button type="submit" disabled={busy}>{busy?"Đang kiểm tra...":"Đăng nhập"} {!busy && <ArrowRight size={17} aria-hidden="true"/>}</Button>
      </form>
      <p className="auth-note">Tài khoản do quản trị viên cấp. Hệ thống không cho phép tự tạo tài khoản học sinh hoặc phụ huynh công khai.</p>
    </div>
  </section>;
}
