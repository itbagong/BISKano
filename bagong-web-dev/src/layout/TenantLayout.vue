<template>
  <div 
    v-if="data.loading"
    class="w-full fixed h-full bg-[#94a3b86b] z-[999] flex justify-center items-center"
  >
    <loader kind="circle"/>
  </div>
  <div
    class="w-full pt-[70px] p-5 justify-center bg-main-bg main-layout"
    :class="[
      data.menuType == 'full'
        ? 'ml-[200px]'
        : data.menuType == 'mini'
        ? 'ml-[50px]'
        : '',
    ]"
  >
    <breadcrumbs :items="layout.breadcrumbs" />
    <RouterView />
  </div>

  <div class="nav_top flex items-center h-[50px] shadow-md">
    <div
      class="w-[200px] h-full flex items-center justify-center gap-2 px-2 bg-white"
      v-if="data.menuType == 'full'"
    >
      <img src="@/assets/img/logo-lg.png" width="80" />
    </div>
    <div
      v-else
      class="w-[50px] h-full flex items-center justify-center bg-white cursor-pointer"
    >
      <img src="@/assets/img/logo-sm.png" width="25" />
    </div>
    <div
      class="grow h-full flex items-center text-black cursor-pointer"
      @click="changeMenuType()"
    >
      <mdicon
        :name="
          data.menuType == 'full'
            ? 'arrow-right'
            : data.menuType == 'mini'
            ? 'menu'
            : 'arrow-left'
        "
        size="22"
        class="text-black"
      />
    </div>
    <div class="flex flex-col pl-2 pr-2">
      <notifcations />
    </div>
    <div class="flex text-black" v-if="auth.appToken != ''">
      <div class="flex flex-col pl-2 pr-2">
        <div class="">{{ auth.appData.DisplayName }}</div>
        <div class="text-[10px]">
          {{ auth.appData.TenantName }}
        </div>
      </div>
    </div>
    <div v-else class="grow h-full flex flex-col px-2">&nbsp;</div>
    <div
      class="mr-4 flex gap-2 h-full items-center justify-center"
      v-if="auth.appToken == ''"
    >
      <mdicon
        size="18"
        name="login"
        class="nav_right_btn"
        @click="router.push('/iam/Login')"
      />
    </div>
    <div class="mr-4 flex gap-1 items-center justify-center" v-else>
      <user-context-menu :auth="auth" />
    </div>
  </div>

  <div
    v-if="data.menuType != 'hide'"
    class="nav_left flex-col gap-1 text-slate-100 divide-y-2 text-black"
    :class="[data.menuType == 'full' ? 'w-[200px]' : 'w-[50px]']"
  >
    <div class="px-3 py-4"> 
      <context-company
        :view-type="data.menuType"
        @change="data.showContextCompany = true"
      /> 
    </div>
    <div class="flex flex-col pt-2">
      <div v-if="data.loadingeMenu" class="flex flex-col p-2">
        <loader v-for="index in 5" :key="index" kind="skeleton" skeleton-kind="input" /> 
      </div>
      <context-menu
        v-else
        as="div"
        v-if="auth.appToken != '' || true"
        v-for="menu in data.appMenu"
        class="w-full px-3 py-2 cursor-pointer hover:bg-slate-200"
        :icon="menu.icon"
        :label="menu.label"
        :url="menu.url"
        :submenu="menu.submenu"
        :view-type="data.menuType"
        :active-menus="activeMenus"
      />
    </div>
  </div>
  <select-company
    v-if="data.showContextCompany"
    :items="data.company"
    @close="data.showContextCompany = false"
    @submit="changeCompany"
  />
</template>

<script setup>
import { useRouter } from "vue-router";
import { authStore } from "@/stores/auth";
import { layoutStore } from "@/stores/layout";
import SelectCompany from "./widgets/SelectCompany.vue";
import ContextCompany from "./widgets/ContextCompany.vue";
import UserContextMenu from "./widgets/UserContextMenu.vue";
import Notifcations from "./widgets/Notifications.vue";
import { reactive, onMounted, watch, computed, inject } from "vue";
import ContextMenu from "@/components/common/ContextMenu.vue";
import Breadcrumbs from "@/components/common/Breadcrumbs.vue";
import { useRoute } from "vue-router";
import Loader from "@/components/common/Loader.vue";
import { util } from "suimjs"
// import appMenu from "@/data/appmenu";

const router = useRouter();
const auth = authStore();
const route = useRoute();
const layout = layoutStore();
const homeUrl = import.meta.env.VITE_HOME_URL;
const axios = inject("axios");

const data = reactive({
  menuType: "mini",
  showContextCompany: false,
  masterData: [],
  appMenu: [],
  mapAppMenu:{},
  menuMasters: [],
  IDmenuMaster: "MasterData",
  company:[],
  loadingeMenu:false,
  loading:false,
});

function changeMenuType() {
  switch (data.menuType) {
    case "full":
      data.menuType = "mini";
      break;

    case "mini":
      data.menuType = "hide";
      break;

    default:
      data.menuType = "full";
  }
}

function generateBreadcrumb(route) { 
   let el = data.mapAppMenu[route.fullPath]
   let breadcrumbs = []
   let end = true


   while(end){
    if(el == undefined){
      end = false
    }else{
      breadcrumbs.unshift({ label: el.Label, url: el.Uri })
      el = data.mapAppMenu[el.ParentMenuID]
    }
   }
   

  layout.setBreadCrumbs(breadcrumbs);
}

const activeMenus = computed({
  get() {
    return layout.breadcrumbs.map((e) => e.label);
  },
});

const addDataMaster = computed({
  get() {
    return layout.addDataMaster;
  },
});

watch(
  route,
  (to) => {
    generateBreadcrumb(to);
  },
  { flush: "pre", immediate: true, deep: true }
);

watch(
  addDataMaster,
  (add) => {
    if (add) {
      init();
    }
  },
  { flush: "pre", immediate: true, deep: true }
);

async function init() {
  if (auth.appToken == "") {
    //return router.push("/sign/in");
  } else {
    data.loading = true
    await fetchCompany()
 
    if(auth.appData?.CompanyID != '' &&  auth.appData?.CompanyID != undefined) {
      data.loading = false
      getMenus() 
      return
    }

    if(data.company.length == 1){
      changeCompany(data.company[0]._id)
    }else{
      data.showContextCompany =  true
      data.loading = false
    } 
    
   
  }
}
async function changeCompany(companyID) {
  data.loading = true
  const url = "/iam/change-data"
  const param = {
      Scope: "BOTH",
      Token: auth.appToken,
      Data: { CompanyID: companyID },
  }
  try {
    const r = await axios.post(url, param)
    auth.setCompanyId(data.selected);
    auth.updateJwt(r.data)
  } catch (e) {
    util.showError(e)
  } finally {
    getMenus()
    data.loading = false
  }
}
async function fetchCompany(){
  data.company = []
  try {
    const r = await axios.post("/tenant/company/gets", {})
    data.company = [...r.data.data]
  } catch (e) {
    util.showError(e)
  }
}
function generateMapMenu(dt){
  return dt.reduce(function(map, obj) {
            const key = obj.Uri !== '' ? obj.Uri : obj._id
            map[key] = obj
            return map
        }, {});
}
function generateMenu(dt) {

  const groupMenu = Object.groupBy(dt, (obj) => obj.ParentMenuID);
 
  
  function createMenu(el){
    return el.reduce((r, e) => {
      if (e.FeatureID === "" || auth.hasFeatureID(e.FeatureID)) {
      
        let submenu = buildMenu(e._id);

        r.push({
          ...e,
          ...{
            label:e.Label,
            url:e.Uri,
            icon: e.Icon,
            pathLabel: e.PathLabel,
            submenu 
          },
        });
      }
      return r;
    }, []).sort((a,b) => a.Priority - b.Priority);
  }

  function buildMenu(_id) { 
    let arr = groupMenu[_id]
    if (arr === undefined) return [];
 
    return  _id == "" ? createMenu(arr) : Object.entries(Object.groupBy(arr, (obj) =>  obj.Section)).reduce((r,e,i)=>{
      r.push(createMenu(e[1]))
      return r
    },[]).filter(e=>e.length>0)
  }

  const r = buildMenu("");

  return r;
}
 
async function fetchMenuMasters() {
  try {
    const r = await axios.post("/tenant/masterdatatype/gets", {Sort: ["Name"]});
    data.menuMasters = r.data.data.map(e => {
      return {
        ParentMenuID: data.IDmenuMaster,
        Label: e.Name,
        Uri: `/tenant/masterdata?MasterDataTypeID=${e._id}&title=${e.Name}`,
        Section:"",
        FeatureID:""
      };
    });
  } catch (e) {
    data.menuMasters = [];
  }
}
async function getMenus() {
  data.loadingeMenu = true
  await fetchMenuMasters();
  axios
    .post(`/admin/menu/gets?AppID=${layout.appID}`, {
      Sort: ["-PathLabel"],
    })
    .then(
      (r) => {
        const dt = [...r.data.data, ...data.menuMasters]
  
        data.mapAppMenu = generateMapMenu(dt)
        data.appMenu = generateMenu(dt); 

      },
      (err) => {
        data.mapAppMenu = {}
        data.appMenu = [];
      }
    ).finally(()=>{
      generateBreadcrumb(route);
      data.loadingeMenu = false
    })
}
 
onMounted(async () => {
  // initMenu();
  init();
});
</script>

<style>
.nav_top {
  @apply z-[999] fixed w-[100%] h-[50px] text-black bg-white;
  width: 100%;
}

.nav_left {
  @apply fixed h-full top-[50px] bg-white;
}

.nav_right_btn {
  @apply text-black hover:opacity-50 cursor-pointer;
}
</style>
