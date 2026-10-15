"use client";
import {FormEvent,useState} from "react";
import {useRouter} from "next/navigation";
import {KeyRound, ShieldCheck} from "lucide-react";
import {Button} from "@/components/ui/button";

export default function PasswordForm(){
 const router=useRouter();
 const [oldPassword,setOldPassword]=useState("");
 const [newPassword,setNewPassword]=useState("");
 const [confirm,setConfirm]=useState("");
 const [busy,setBusy]=useState(false);
 const [error,setError]=useState("");
 async function submit(e:FormEvent<HTMLFormElement>){
  e.preventDefault();setError("");
  if(newPassword!==confirm){setError("Mật khẩu xác nhận chưa khớp.");return;}
  if(newPassword===oldPassword){setError("Mật khẩu mới cần khác mật khẩu cũ.");return;}
  setBusy(true);
  try{
   const res=await fetch("/api/auth/password",{
    method:"POST",headers:{"content-type":"application/json"},
    body:JSON.stringify({current_password:oldPassword,new_password:newPassword})
   });
   const result=await res.json();
   if(!res.ok)throw new Error(result.error||"Không thể đổi mật khẩu");
   setOldPassword("");setNewPassword("");setConfirm("");
   router.replace("/login?password=changed");router.refresh();
  }catch(err){setError(err instanceof Error?err.message:"Lỗi đổi mật khẩu");}
  finally{setBusy(false);}
 }
 return <section className="auth-card">
  <div className="auth-symbol" aria-hidden="true"><KeyRound size={29}/></div>
  <h2>Đổi mật khẩu</h2>
  <p className="muted">Nhập mật khẩu hiện tại và tạo một mật khẩu mới mạnh (12–72 ký tự ASCII hoặc tối đa 72 byte UTF-8).</p>
  {error&&<div className="form-error" role="alert">{error}</div>}
  <form className="auth-form" onSubmit={submit}>
   <label>Mật khẩu hiện tại<input autoComplete="current-password" type="password" required value={oldPassword} onChange={e=>setOldPassword(e.target.value)}/></label>
   <label>Mật khẩu mới<input autoComplete="new-password" type="password" minLength={12} required value={newPassword} onChange={e=>setNewPassword(e.target.value)}/></label>
   <label>Nhập lại mật khẩu mới<input autoComplete="new-password" type="password" minLength={12} required value={confirm} onChange={e=>setConfirm(e.target.value)}/></label>
   <Button disabled={busy||newPassword.length<12||newPassword!==confirm} type="submit"><ShieldCheck size={18} aria-hidden="true"/>{busy?"Đang cập nhật…":"Lưu và đăng nhập lại"}</Button>
  </form>
  <p className="auth-note">Mật khẩu không được gửi vào nhật ký ứng dụng. Hệ thống chỉ lưu mã băm đã tạo bằng bcrypt.</p>
 </section>;
}
