import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import Link from "next/link";
import ContentEditor from "./editor";

export default async function EditorialPage(){
 const token=(await cookies()).get("openedu_session")?.value;
 if(!token)redirect("/login");
 let roles:string[]=[];let schoolID="";
 try{
  const res=await fetch((process.env.API_INTERNAL_URL||"http://localhost:8080")+"/v1/auth/me",{
   cache:"no-store",signal:AbortSignal.timeout(7000),headers:{Authorization:"Bearer "+token}
  });
  if(!res.ok)redirect("/login");
  const data=await res.json() as {roles:string[];school_id:string};
  roles=data.roles;schoolID=data.school_id||"";
 }catch{redirect("/login");}
 if(!roles.some(role=>["platform_admin","school_admin","content_admin","content_reviewer"].includes(role)))redirect("/dashboard");
 return <section className="container dashboard">
  <div className="dashboard-top"><div><span className="eyebrow section-eyebrow">OPENEDU CONTENT STUDIO</span><h1>Quản trị học liệu</h1><p>Soạn bài và quản lý quy trình duyệt, xuất bản với chứng cứ quyền sử dụng.</p></div><Link className="back" href="/dashboard">← Dashboard</Link></div>
  <ContentEditor global={roles.includes("platform_admin")&&!schoolID}/>
 </section>;
}
