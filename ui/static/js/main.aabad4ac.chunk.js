(this.webpackJsonpclient=this.webpackJsonpclient||[]).push([[0],{117:function(e,t,a){},123:function(e,t,a){"use strict";a.r(t);var n=a(0),r=a.n(n),o=a(9),l=a.n(o),c=(a(98),a(72)),i=a(14),m=a(54),u=a(10),s=(a(99),a(153)),g=a(73),p=a.n(g).a.create({baseURL:""}),h=a(157),f=a(176),d=(a(117),Object(s.a)((function(e){return{searchBox:{position:"relative",width:"100%",height:"50px"},alarm:{position:"absolute",bottom:"0",paddingLeft:"3rem",color:"red"}}})));function E(e){var t=e.onChange,a=e.onKeyPress,n=e.empty,o=d();return r.a.createElement("div",{className:o.searchBox},r.a.createElement("div",{className:"search-container"},r.a.createElement("input",{type:"text",onChange:t,onKeyPress:a,placeholder:"Search your comic"}),r.a.createElement("div",{className:"search"}),n?r.a.createElement("div",{className:o.alarm},"Kh\xf4ng t\xecm th\u1ea5y truy\u1ec7n !!!"):null))}var b=a(167),v=a(175),y=a(155),x=a(156),w=a(158),k=a(159),j=a(77),N=a.n(j),O=a(160),C=Object(s.a)((function(e){return{popup:{display:"flex",justifyContent:"center"},center:{margin:"auto"}}})),S=function(e){var t=e.userID,a=e.comic,o=e.open,l=e.close,c=C(),m=Object(n.useState)(!1),u=Object(i.a)(m,2),s=u[0],g=u[1],f=Object(n.useState)(!1),d=Object(i.a)(f,2),E=d[0],b=d[1];return r.a.createElement(v.a,{open:o,onClose:l},r.a.createElement(y.a,{id:"dialog-title"},"Delete comic"),r.a.createElement(x.a,{className:c.popup},s?r.a.createElement(h.a,null):r.a.createElement("div",null,E?r.a.createElement(N.a,null):r.a.createElement("div",null,r.a.createElement(w.a,{id:"alert-dialog-description"},"Remove notification when"," ",r.a.createElement("span",{style:{fontWeight:"bold"}},a.name)," ","update new chapter"),r.a.createElement(k.a,null,r.a.createElement(O.a,{variant:"contained",color:"secondary",onClick:l},"Cancel"),r.a.createElement(O.a,{variant:"contained",color:"primary",onClick:function(e){e.preventDefault(),g(!0),b(!1),p.delete("api/v1/users/".concat(t,"/comics/").concat(a.id)).then((function(e){g(!1),b(!0),window.location.reload()})).catch((function(e){return console.log(e)}))}},"OK"))))))},I=a(162),D=a(161),B=a(164),T=a(163),L=a(165),F=a(166),_=a(57),W=a(78),R=a.n(W),z=Object(s.a)((function(e){return{root:{display:"flex",flexFlow:"column",maxWidth:"400px",height:"350px",margin:"auto"},media:{backgroundColor:"gray",margin:"auto",width:"400px",height:"200px"},comicInfo:{overflow:"hidden",textOverflow:"ellipsis"},noMaxWidth:{maxWidth:"none"}}}));function A(e){e._;var t=e.userID,a=e.comic,o=z(),l=Object(n.useState)(!1),c=Object(i.a)(l,2),m=c[0],u=c[1];return r.a.createElement(D.a,{className:o.root},r.a.createElement(I.a,{href:a.chapURL,target:"_blank",rel:"noopener noreferrer"},r.a.createElement(T.a,{className:o.media,component:"img",alt:a.name,image:a.imgURL})),r.a.createElement(B.a,{className:o.content},r.a.createElement(_.a,{className:o.comicInfo,component:"div"},r.a.createElement(I.a,{href:a.url,target:"_blank",rel:"noopener noreferrer",style:{color:"inherit",fontSize:"1.5rem",whiteSpace:"nowrap"}},a.name)),r.a.createElement(_.a,{className:o.comicInfo,component:"div",color:"textSecondary"},r.a.createElement(I.a,{href:a.chapURL,target:"_blank",rel:"noopener noreferrer",style:{marinTop:"5px",color:"inherit",fontSize:"0.75rem",whiteSpace:"nowrap"}},a.latestChap))),r.a.createElement(L.a,{disableSpacing:!0,style:{padding:"5px"}},r.a.createElement(_.a,{style:{color:"\t#808080",marginLeft:"10px"}},a.page),r.a.createElement(F.a,{className:o.delete,"aria-label":"Delete",onClick:function(e){e.preventDefault(),u(!0)},style:{marginLeft:"auto"}},r.a.createElement(R.a,null))),r.a.createElement(S,{userID:t,comic:a,open:m,close:function(e){e.preventDefault(),u(!1)}}))}function K(e){var t=e.page,a=e.limit,n=e.comics,o=e.userID;return r.a.createElement(b.a,{container:!0,spacing:2},n.slice((t-1)*a,t*a).map((function(e){return r.a.createElement(b.a,{key:e.id,item:!0,xs:12,sm:12,md:6,lg:4,xl:4},r.a.createElement(A,{userID:o,comic:e}))})))}var q=Object(s.a)((function(e){return{container:{width:"60%",margin:"0 auto"},page:{marginTop:"60px"},pagination:{margin:"20px",paddingBottom:"10px",display:"flex",justifyContent:"center"},spinnerContainer:{width:"100%",display:"flex",justifyContent:"center",paddingTop:"2rem"}}})),U=function(e){var t=q(),a=Object(n.useState)(!1),o=Object(i.a)(a,2),l=o[0],c=o[1],m=Object(n.useState)(""),u=Object(i.a)(m,2),s=u[0],g=u[1],d=Object(n.useState)([]),b=Object(i.a)(d,2),v=b[0],y=b[1],x=Object(n.useState)(1),w=Object(i.a)(x,2),k=w[0],j=w[1],N=Object(n.useState)(!1),O=Object(i.a)(N,2),C=O[0],S=O[1],I=function(e){p.get("api/v1/users/".concat(e,"/comics")).then((function(e){y(e.data.comics),S(!0),c(!1)})).catch((function(e){return console.log(e)}))},D=function(){if(""===s)return j(1),void I(e.userID);p.get("api/v1/users/".concat(e.userID,"/comics"),{params:{q:s}}).then((function(e){g(""),0!==e.data.comics.length?y(e.data.comics):c(!0)})).catch((function(e){return console.log(e)}))};return Object(n.useEffect)((function(){I(e.userID)}),[e.userID]),r.a.createElement("div",{className:t.container},C?r.a.createElement(E,{value:s,onChange:function(e){g(e.target.value)},onClick:D,onKeyPress:function(e){"Enter"===e.key&&(e.target.value="",D())},empty:l}):null,C?r.a.createElement("div",{className:t.page},r.a.createElement("div",null,0!==v.length?r.a.createElement("div",null,r.a.createElement(K,{page:k,limit:6,comics:v,userID:e.userID}),r.a.createElement("div",{className:t.pagination},r.a.createElement(f.a,{count:Math.ceil(v.length/6),onChange:function(e,t){j(t)}}))):r.a.createElement("h2",null,"B\u1ea1n ch\u01b0a \u0111\u0103ng k\xfd nh\u1eadn th\xf4ng b\xe1o cho truy\u1ec7n, xem h\u01b0\u1edbng d\u1eabn t\u1ea1i \u0111\xe2y"))):r.a.createElement("div",{className:t.spinnerContainer},r.a.createElement(h.a,{color:"inherit"})))},M=(Object(s.a)({form:{display:"flex",height:"50px"}}),Object(s.a)((function(e){return{root:{height:"100%",backgroundColor:"#F5F5F5"}}}))),P=function(e){var t=e.userID,a=M();return r.a.createElement("div",{className:a.root},r.a.createElement("div",{className:a.form}),t?r.a.createElement(U,{userID:t}):null)},H=a(41),J=Object(s.a)((function(e){return{root:{heigth:"100%",display:"flex",backgroundImage:'url("'.concat("/assets/bg-3.png",'")'),justifyContent:"center"},container:{height:"100vh",display:"flex",flexWrap:"no-wrap",width:"80%"},login:Object(H.a)({display:"flex",alignItems:"center",flexDirection:"column",width:"60%"},e.breakpoints.down("sm"),{width:"100%"}),banner:{backgroundColor:"rgba(255,255,255,0.8)",borderRadius:"2%",margin:"auto 0",fontFamily:"Itim",display:"flex",flexDirection:"column",alignItems:"center"},logo:{margin:"auto"},botName:{fontSize:"3rem",textAlign:"center"},loginButton:{fontSize:"2.5rem",margin:"40px",transition:"0.3s","&:hover":{backgroundColor:"black",color:"white"}},slider:Object(H.a)({display:"flex",width:"60%",justifyContent:"center",alignItems:"center"},e.breakpoints.down("sm"),{display:"none"})}}));function G(){var e=J();return r.a.createElement("div",{className:e.root},r.a.createElement("div",{className:e.container},r.a.createElement("div",{className:e.login},r.a.createElement("div",{className:e.banner},r.a.createElement("img",{className:e.logo,src:"/assets/chatbot.svg",alt:"logo",style:{height:"auto",width:"100%"}}),r.a.createElement("div",{className:e.botName},"COMIC NOTIFY BOT"),r.a.createElement("button",{className:e.loginButton,onClick:function(){window.location.href="https://www.facebook.com/v8.0/dialog/oauth?client_id=".concat("145792170193635","&redirect_uri=").concat("https://comicnotifier.herokuapp.com/auth","&state=").concat("quangmt2")}},"Log in with Facebook"))),r.a.createElement("div",{className:(e.content,e.slider)},r.a.createElement("img",{src:"/assets/bot-logo.jpg",alt:"slider",style:{height:"auto",width:"60%"}}))))}var X=a(171),Y=a(172),$=a(174),Q=a(177),V=a(170),Z=a(168),ee=a(169),te=a(82),ae=a(179),ne=a(178),re=a(79),oe=a.n(re),le=Object(s.a)((function(e){return{root:{display:"flex"},paper:{marginRight:e.spacing(2)}}}));function ce(e){var t=e.toAbout,a=e.toTutorial,n=e.logout,o=le(),l=r.a.useState(!1),c=Object(i.a)(l,2),m=c[0],u=c[1],s=r.a.useRef(null),g=function(e){s.current&&s.current.contains(e.target)||u(!1)},p=function(e){g(e),t()},h=function(e){g(e),a()},f=function(e){g(e),n()};function d(e){"Tab"===e.key&&(e.preventDefault(),u(!1))}var E=r.a.useRef(m);return r.a.useEffect((function(){!0===E.current&&!1===m&&s.current.focus(),E.current=m}),[m]),r.a.createElement("div",{className:o.root},r.a.createElement("div",null,r.a.createElement($.a,{ref:s,"aria-controls":m?"menu-list-grow":void 0,"aria-haspopup":"true",onClick:function(){u((function(e){return!e}))},m:2},r.a.createElement(oe.a,null)),r.a.createElement(Z.a,{open:m,anchorEl:s.current,role:void 0,transition:!0,disablePortal:!0},(function(e){var t=e.TransitionProps,a=e.placement;return r.a.createElement(ee.a,Object.assign({},t,{style:{transformOrigin:"bottom"===a?"center top":"center bottom"}}),r.a.createElement(te.a,null,r.a.createElement(V.a,{onClickAway:g},r.a.createElement(ae.a,{autoFocusItem:m,id:"menu-list-grow",onKeyDown:d},r.a.createElement(ne.a,{onClick:p},"About"),r.a.createElement(ne.a,{onClick:h},"Tutorial"),r.a.createElement(ne.a,{onClick:f},"Logout")))))}))))}var ie=Object(s.a)({appbar:{background:"white",color:"#000",fontFamily:"Nunito",position:"sticky"},appbarWrapper:{width:"60%",margin:"0 auto",padding:"0",minWidth:"360px"},appbarTitle:{display:"flex",flex:"1"},appbarName:{padding:"auto",fontSize:"1.5rem",margin:"auto 0"},logoutButton:{fontFamily:"Nunito",fontSize:"0.75rem"},avatar:{borderRadius:"50%"}}),me=function(e){var t=e.user,a=e.clearLocalStorage,n=(t.psid,t.name,t.profile_pic),o=ie(),l=Object(u.f)(),c=function(){p.get("/logout").then((function(e){a(),l.push("/")})).catch((function(e){return console.log(e)}))},i=function(){l.push("/about")},m=function(){l.push("/tutorial")};return r.a.createElement(X.a,{className:o.appbar},r.a.createElement(Y.a,{className:o.appbarWrapper},r.a.createElement("div",{className:o.appbarTitle,onClick:function(){l.push("/")},style:{cursor:"pointer"}},r.a.createElement("img",{src:"/assets/chatbot.svg",alt:"logo",style:{height:"auto",width:"75px"}}),r.a.createElement("div",{className:o.appbarName},"Comic Notify")),r.a.createElement($.a,{display:"flex",fontWeight:"fontWeightLight",alignItems:"center"},r.a.createElement(Q.a,{only:["xs","sm"]},r.a.createElement($.a,{mr:2},r.a.createElement(O.a,{className:o.logoutButton,onClick:m},"Tutorial"),r.a.createElement(O.a,{className:o.logoutButton,onClick:i},"About"))),r.a.createElement(Q.a,{mdUp:!0},r.a.createElement(ce,{toAbout:i,toTutorial:m,logout:c})),r.a.createElement(Q.a,{only:["xs","sm"]},r.a.createElement("img",{className:o.avatar,src:n,alt:"",style:{height:"auto",width:"35px"}}),r.a.createElement(O.a,{className:o.logoutButton,onClick:c},"Log Out")))))},ue=Object(s.a)((function(e){return{root:{"& li":{marginBottom:"25px"},fontFamily:"Nunito",width:"60%",margin:"60px auto",paddingBottom:"200px",fontSize:"1.2rem"}}}));function se(){var e=ue();return r.a.createElement("div",{className:e.root},r.a.createElement("h3",null,"Chatbot:"," ",r.a.createElement("a",{href:"https://m.me/Cominify",target:"_blank",rel:"noopener noreferrer"},"m.me/Cominify")),r.a.createElement("p",null,"Hi\u1ec7n t\u1ea1i ch\u1ec9 h\u1ed7 tr\u1ee3 3 page:"," ",r.a.createElement("a",{href:"https://beeng.net",target:"_blank",rel:"noopener noreferrer"},"beeng.net"),",",r.a.createElement("a",{href:"https://blogtruyen.vn",target:"_blank",rel:"noopener noreferrer"}," ","blogtruyen.vn"),","," ",r.a.createElement("a",{href:"https://truyendep.com",target:"_blank",rel:"noopener noreferrer"},"truyendep.com")," ","(a.k.a mangaK)"),r.a.createElement("br",null),r.a.createElement("h3",null,"H\u01b0\u1edbng d\u1eabn \u0111\u0103ng k\xed truy\u1ec7n"),r.a.createElement("ul",{style:{listStyleType:"none",padding:"0"}},r.a.createElement("li",null,r.a.createElement("div",null,"1. L\u1ea5y link truy\u1ec7n mu\u1ed1n \u0111\u0103ng k\xed, v\xed d\u1ee5:"),r.a.createElement("a",{href:"https://beeng.net/truyen-tranh-online/dao-hai-tac-31953",target:"_blank",rel:"noopener noreferrer"},"https://beeng.net/truyen-tranh-online/dao-hai-tac-31953")),r.a.createElement("li",null,r.a.createElement("div",null,"2. Copy \u0111\u01b0\u1eddng d\u1eabn:"),r.a.createElement("img",{src:"/assets/tutor-1.jpg",alt:"",style:{width:"50%"}})),r.a.createElement("li",null,r.a.createElement("div",null,"3. G\u1edfi cho BOT"),r.a.createElement("img",{src:"/assets/tutor-2.jpg",alt:"",style:{width:"50%"}})),r.a.createElement("li",null,r.a.createElement("div",null,"4. \u0110\u0103ng k\xed th\xe0nh c\xf4ng, BOT s\u1ebd g\u1edfi l\u1ea1i chap m\u1edbi nh\u1ea5t"),r.a.createElement("img",{src:"/assets/tutor-3.jpg",alt:"",style:{width:"50%"}})),r.a.createElement("li",null,r.a.createElement("p",null,"5. Xong, gi\u1edd th\xec... ng\u1ed3i rung \u0111\xf9i th\xf4i :)"),r.a.createElement("p",{style:{textAlign:"justify"}},"- Khi n\xe0o c\xf3 chap m\u1edbi, BOT s\u1ebd g\u1edfi v\u1ec1 th\xf4ng b\xe1o qu\xe1 Mesenger"),r.a.createElement("p",{style:{textAlign:"justify"}},"- N\u1ebfu kh\xf4ng mu\u1ed1n nh\u1eadn th\xf4ng b\xe1o khi truy\u1ec7n update n\u1eefa, nh\u1ea5n n\xfat"," ",r.a.createElement("span",{style:{fontSize:"1.5rem"}},"Unsubscribe")," tr\xean message"))))}var ge=Object(s.a)((function(e){return{root:{"& li":{marginBottom:"25px"},fontFamily:"Nunito",width:"60%",margin:"60px auto",paddingBottom:"200px",fontSize:"1.2rem"}}}));function pe(){var e=ge();return r.a.createElement("div",{className:e.root},r.a.createElement("p",null," ","N\u1ebfu c\xf3 l\u1ed7i g\xec, hay c\xf3 \u0111\u1ec1 xu\u1ea5t g\xec inbox cho m\xecnh qua Facebook:"," ",r.a.createElement("a",{href:"https://www.facebook.com/thienquang2804/",target:"_blank",rel:"noopener noreferrer"},"https://www.facebook.com/thienquang2804/")),r.a.createElement("p",null,"C\xf3 th\u1eddi gian s\u1ebd fix :)"))}var he=function(){var e=Object(n.useState)({}),t=Object(i.a)(e,2),a=t[0],o=t[1];var l=function(){var e=function(e){for(var t=e+"=",a=document.cookie.split(";"),n=0;n<a.length;n++){for(var r=a[n];" "===r.charAt(0);)r=r.substring(1,r.length);if(0===r.indexOf(t))return r.substring(t.length,r.length)}return null}("upid");e&&p.get("/api/v1/users/".concat(e)).then((function(e){o(Object(c.a)({},e.data))})).catch((function(e){return console.log(e)}))};return Object(n.useEffect)((function(){p.get("/status").then((function(e){localStorage.setItem("logged","true")})).catch((function(e){localStorage.removeItem("logged"),o({})})),l()}),[]),r.a.createElement(m.a,null,localStorage.getItem("logged")?r.a.createElement(me,{user:a,clearLocalStorage:function(){localStorage.removeItem("logged"),o({})}}):null,r.a.createElement(u.c,null,r.a.createElement(u.a,{path:"/",exact:!0},localStorage.getItem("logged")?r.a.createElement(P,{userID:a.appid}):r.a.createElement(G,null)),r.a.createElement(u.a,{path:"/about"},r.a.createElement(pe,null)),r.a.createElement(u.a,{path:"/tutorial"},r.a.createElement(se,null))))},fe=a(173);Boolean("localhost"===window.location.hostname||"[::1]"===window.location.hostname||window.location.hostname.match(/^127(?:\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$/));l.a.render(r.a.createElement(r.a.StrictMode,null,r.a.createElement(fe.a,null),r.a.createElement(he,null)),document.getElementById("root")),"serviceWorker"in navigator&&navigator.serviceWorker.ready.then((function(e){e.unregister()})).catch((function(e){console.error(e.message)}))},93:function(e,t,a){e.exports=a(123)},98:function(e,t,a){},99:function(e,t,a){}},[[93,1,2]]]);
//# sourceMappingURL=main.aabad4ac.chunk.js.map