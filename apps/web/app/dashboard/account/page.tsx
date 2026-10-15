import Link from "next/link";
import {cookies} from "next/headers";
import {redirect} from "next/navigation";
import { ArrowLeft, KeyRound } from "lucide-react";
import PasswordForm from "./password-form";
export default async function AccountSettings(){
 if(!(await cookies()).get("openedu_session")?.value)redirect("/login");
 return <section className="container dashboard">
  <div className="dashboard-top"><div>
   <span className="eyebrow section-eyebrow">OPENEDU ACCOUNT SECURITY</span>
   <h1 className="icon-heading"><KeyRound size={28} aria-hidden="true"/>Bảo mật tài khoản</h1>
   <p>Thay đổi mật khẩu để bảo vệ tài khoản. Sau khi thay đổi, các phiên đăng nhập cũ sẽ bị vô hiệu hóa.</p>
  </div><Link href="/dashboard" className="back"><ArrowLeft size={16} aria-hidden="true"/>Quay lại</Link></div>
  <div className="auth-wrap account-form-wrap"><PasswordForm/></div>
 </section>;
}
