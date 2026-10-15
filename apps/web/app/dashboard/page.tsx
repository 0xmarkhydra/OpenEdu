import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import Link from "next/link";
import AdminPanel from "./panel";
import SchoolPanel from "./school-panel";
import ParentPanel from "./parent-panel";
import { ArrowLeft, BookOpen, LogOut } from "lucide-react";
import { Button } from "@/components/ui/button";

type Profile={user_id:string;school_id?:string;roles:string[]};
type School={id:string;name:string;slug:string};

async function api<T>(route:string,token:string):Promise<T|null>{
  try{
    const res=await fetch((process.env.API_INTERNAL_URL||"http://localhost:8080")+route,{
      cache:"no-store",signal:AbortSignal.timeout(6000),headers:{Authorization:"Bearer "+token}
    });
    return res.ok ? (await res.json() as T) : null;
  }catch{return null;}
}

export default async function DashboardPage(){
  const token=(await cookies()).get("openedu_session")?.value;
  if(!token) redirect("/login");
  const profile=await api<Profile>("/v1/auth/me",token);
  if(!profile) redirect("/login");
  const isPlatform=profile.roles.includes("platform_admin") && !profile.school_id;
  const data=isPlatform ? await api<{items:School[]}>("/v1/admin/schools",token) : null;
  return <section className="container dashboard">
    <div className="dashboard-top">
      <div><span className="eyebrow section-eyebrow">OPENEDU WORKSPACE</span><h1>Trung tâm quản trị</h1><p>Vai trò: <strong>{profile.roles.join(", ")}</strong>{profile.school_id && <> · Trường: {profile.school_id}</>}</p></div>
      <div className="dashboard-tools"><Link className="back" href="/"><ArrowLeft size={16} aria-hidden="true"/>Thư viện</Link>{profile.roles.some(r=>["platform_admin","school_admin","content_admin","content_reviewer"].includes(r))&&<Link className="back" href="/dashboard/content"><BookOpen size={16} aria-hidden="true"/>Quản trị học liệu</Link>}<form action="/api/auth/logout" method="post"><LogoutButton/></form></div>
    </div>
    {isPlatform ? <AdminPanel initialSchools={data?.items||[]}/> : profile.school_id ? profile.roles.includes("guardian") && !profile.roles.some(r=>["school_admin","principal","teacher","student"].includes(r)) ? <ParentPanel/> : <SchoolPanel schoolId={profile.school_id} userRoles={profile.roles}/> : <div className="empty-state"><h3>Chưa có trường liên kết</h3><p>Liên hệ quản trị viên để được cấp quyền truy cập trường học.</p></div>}
  </section>;
}

function LogoutButton(){
  return <Button variant="outline" className="logout-btn" type="submit"><LogOut size={16} aria-hidden="true"/>Đăng xuất</Button>;
}