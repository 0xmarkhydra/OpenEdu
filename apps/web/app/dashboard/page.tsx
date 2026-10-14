import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import Link from "next/link";
import AdminPanel from "./panel";

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
      <div className="dashboard-tools"><Link className="back" href="/">← Thư viện</Link><form action="/api/auth/logout" method="post"><LogoutButton/></form></div>
    </div>
    {isPlatform ? <AdminPanel initialSchools={data?.items||[]}/> : <div className="empty-state"><h3>Đã xác thực tài khoản</h3><p>Những thao tác của trường, giáo viên và phụ huynh sẽ được mở theo phân quyền sau khi chức năng hoàn thiện.</p></div>}
  </section>;
}

function LogoutButton(){
  return <button className="logout-btn" type="submit">Đăng xuất</button>;
}
