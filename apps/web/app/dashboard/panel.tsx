"use client";

import { FormEvent, useState } from "react";
type School={id:string;name:string;slug:string};
type Result={error?:string;id?:string};

export default function AdminPanel({initialSchools}:{initialSchools:School[]}){
 const [schools,setSchools]=useState(initialSchools);
 const [name,setName]=useState("");
 const [slug,setSlug]=useState("");
 const [selected,setSelected]=useState(initialSchools[0]?.id||"");
 const [memberEmail,setMemberEmail]=useState("");
 const [memberName,setMemberName]=useState("");
 const [memberPassword,setMemberPassword]=useState("");
 const [role,setRole]=useState("school_admin");
 const [notice,setNotice]=useState("");
 const [busy,setBusy]=useState(false);
 async function createSchool(e:FormEvent<HTMLFormElement>){
  e.preventDefault();setBusy(true);setNotice("");
  try{
   const r=await fetch("/api/admin/schools",{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({name,slug})});
   const data=await r.json() as School & Result;
   if(!r.ok){setNotice(data.error||"Không thể tạo trường");return;}
   setSchools(old=>[...old,data]);setSelected(data.id);setName("");setSlug("");setNotice("Đã tạo trường học thành công.");
  }catch{setNotice("Không thể kết nối đến API");}
  finally{setBusy(false)}
 }
 async function createMember(e:FormEvent<HTMLFormElement>){
  e.preventDefault();setBusy(true);setNotice("");
  try{
   if(!selected){setNotice("Chọn trường trước khi tạo thành viên.");return;}
   const r=await fetch("/api/admin/schools/"+encodeURIComponent(selected)+"/members",{method:"POST",headers:{"content-type":"application/json"},
    body:JSON.stringify({email:memberEmail,name:memberName,password:memberPassword,role})});
   const data=await r.json() as Result;
   if(!r.ok){setNotice(data.error||"Không thể tạo tài khoản");return;}
   setMemberEmail("");setMemberName("");setMemberPassword("");setNotice("Đã cấp tài khoản "+role+" vào trường.");
  }catch{setNotice("Không thể kết nối đến API");}
  finally{setBusy(false)}
 }
 return <div className="admin-grid">
  <section className="admin-card">
   <h2>Danh sách trường học</h2>
   <p>Quản lý các đơn vị sử dụng OpenEdu. Mỗi trường có không gian dữ liệu riêng.</p>
   <div className="school-list">{schools.length?schools.map(s=><div className="school-row" key={s.id}><div><strong>{s.name}</strong><small>{s.slug}</small></div><button onClick={()=>setSelected(s.id)} type="button" className={selected===s.id?"school-selected":""}>Chọn</button></div>):<p>Chưa có trường nào. Tạo trường đầu tiên bên dưới.</p>}</div>
   <form className="auth-form" onSubmit={createSchool}>
    <h3>Thêm trường học</h3>
    <label>Tên trường<input required minLength={2} maxLength={160} value={name} onChange={e=>setName(e.target.value)} placeholder="Trường THCS ..."/></label>
    <label>Mã nhận diện (slug)<input required minLength={3} pattern="[a-z0-9][a-z0-9-]{1,48}[a-z0-9]" value={slug} onChange={e=>setSlug(e.target.value.toLowerCase())} placeholder="truong-thcs-abc"/></label>
    <button disabled={busy}>Tạo trường →</button>
   </form>
  </section>
  <section className="admin-card">
   <h2>Thêm thành viên</h2>
   <p>Cấp tài khoản vào trường đã chọn. Chỉ quản trị viên nền tảng được thực hiện.</p>
   <form className="auth-form" onSubmit={createMember}>
    <label>Trường<select value={selected} onChange={e=>setSelected(e.target.value)} required>{schools.length===0?<option value="">Chưa có trường</option>:schools.map(s=><option key={s.id} value={s.id}>{s.name}</option>)}</select></label>
    <label>Họ tên<input required minLength={2} maxLength={160} value={memberName} onChange={e=>setMemberName(e.target.value)}/></label>
    <label>Email<input type="email" required value={memberEmail} onChange={e=>setMemberEmail(e.target.value)}/></label>
    <label>Vai trò<select value={role} onChange={e=>setRole(e.target.value)}>
     <option value="school_admin">Quản trị trường</option><option value="principal">Ban giám hiệu</option><option value="teacher">Giáo viên</option>
     <option value="student">Học sinh</option><option value="guardian">Phụ huynh</option><option value="content_admin">Quản lý học liệu</option><option value="content_reviewer">Duyệt học liệu</option>
    </select></label>
    <label>Mật khẩu khởi tạo<input type="password" autoComplete="new-password" required minLength={12} maxLength={72} value={memberPassword} onChange={e=>setMemberPassword(e.target.value)}/></label>
    <button disabled={busy||!selected}>Cấp tài khoản →</button>
   </form>
  </section>
  {notice&&<p className="admin-notice" role="status">{notice}</p>}
 </div>;
}
