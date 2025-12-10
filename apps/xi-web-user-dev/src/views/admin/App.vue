<template>
    <div class="w-full">
      <data-list 
        class="card"
        ref="dataCtl" title="Application" grid-config="/admin/app/gridconfig" form-config="/admin/app/formconfig"
        grid-read="/admin/app/gets" form-read="/admin/app/get" grid-mode="grid" grid-delete="/admin/app/delete"
        form-insert="/admin/app/insert" form-update="/admin/app/update" :grid-fields="['Enable']" 
        grid-hide-select
        :init-app-mode="data.appMode" :init-form-mode="data.formMode" :form-tabs-edit="['General', 'Menu']"
        stay-on-form-after-save @formNewData="newRole" @formEditData="formOpen" @alterGridConfig="alterGridConfig">
        <template #grid_Enable="{ item }">
          <mdicon v-if="item.Enable" class="text-primary" size="16" name="check-bold" />
          <mdicon v-else class="text-error" size="16" name="close-thick" />
        </template>
        <template #form_tab_Menu="{ item }">
          <div class="flex gap-5">
              <div class="w-[200px] mt-[3px]">
                <s-input label="View As" 
                  v-model="data.viewAs"
                  lookup-url="/iam/user/find" lookup-key="_id" :lookup-labels="['DisplayName']" use-list />
                 
                   
                <context-menu
                  as="div"
                  v-for="menu in data.menuData"
                  class="w-full py-2 cursor-pointer hover:bg-nav-title-b"
                  :icon="menu.icon"
                  :label="menu.label"
                  :url="menu.url"
                  :submenu="menu.submenu"
                  left="left-[130px]"
                  view-type="full"
                />
              </div>
              <div class="w-full">
                <data-list no-gap grid-hide-select grid-sort-field="PathLabel" grid-sort-direction="asc"
                  grid-mode="grid" form-config-new="/admin/menu/new/formconfig" form-config="/admin/menu/formconfig"
                  grid-config="/admin/menu/gridconfig" :grid-read="`/admin/menu/gets?AppID=${item._id}`"
                  form-insert="/admin/menu/insert"  form-update="/admin/menu/update" init-form-mode="edit" 
                  @form-new-data="newMenu" @control-mode-changed="ctlModeChanged"
                >
                </data-list>
              </div>
          </div>
        </template>
      </data-list>
    </div>
  </template>
  
  <script setup>
  import { reactive, ref, nextTick, inject } from "vue";
  import { layoutStore } from "@/stores/layout.js";
  import { DataList, SInput, util } from "suimjs";
  import ContextMenu from "@/components/common/ContextMenu.vue";
  import { authStore } from "@/stores/auth.js";
  authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})

  layoutStore().name = "tenant";

  
  
  const dataCtl = ref(null);

  const axios = inject("axios");
  
  const data = reactive({
    viewAs: "",
    appMode: "grid",
    formMode: "edit",
    menuData: []
  });
  
  function alterGridConfig (cfg) {
    //cfg.fields = cfg.fields.filter(f => f.field!="TenantName")
  }
  
  function newRole(record) {
    record.Name = "";
    record.Enable = true;
  
    formOpen(record)
  }

  function newMenu(record) {
    const app = dataCtl.value.getFormRecord();
    record.AppID = app._id;
  }

  function ctlModeChanged(newMode) {
    if (newMode=="grid") {
      data.menuData = [];
      const app = dataCtl.value.getFormRecord();
      reloadMenu(app._id)
    }
  }

  function reloadMenu(appID) {
    axios.post(`/admin/menu/find?AppID=${appID}`).then(r => {
      console.log(r)
      data.menuData = buildMenuFromDB(r.data,"").flat();
      console.log( data.menuData)
    }, e => util.showWarning(e))
  }

  function buildMenuFromDB(records, parentID, assignUrl) {
    let res = records.
      filter(m => m.ParentMenuID==parentID).
      map(m => {
        return {
          _id: m._id,
          label: m.Label,
          icon: m.Icon,
          url:  assignUrl==undefined ? m.Uri : assignUrl,
          expand: m.Expand,
          priority: m.Priority
        }
      }).sort((a,b) => a.priority - b.priority);
    
    res = res.map(re => {
      const subs = buildMenuFromDB(records, re._id);
      if (subs.length > 0) {
        if (re.expand) {
          return subs;
        } else {
          re.submenu = subs;
          return re;
        }
      }

      return re;
    });

    return res;
  }
  
  function formOpen(record) {
    nextTick(() => {
      nextTick(() => {
        if (record._id != "") {
          reloadMenu(record._id);
        }

        //dataCtl.value.setFormFieldAttr("TenantID", "hide", true)
        if (record._id!=undefined && record._id!=null && record._id!="") dataCtl.value.setFormFieldAttr("_id", "readOnly", true)
          else dataCtl.value.setFormFieldAttr("_id", "readOnly", false)
        dataCtl.value.setFormFieldAttr("_id", "rules", [
          (v) => {
            let vLen = 0;
            let consistsInvalidChar = false;
  
            v.split("").forEach((ch) => {
              vLen++;
              const validCar =
                "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstyuvxxyz_-".indexOf(ch) >= 0;
              if (!validCar) consistsInvalidChar = true;
            });
  
            if (vLen < 4 || consistsInvalidChar)
              return "minimal length is 4 and alphabet only";
            return "";
          },
        ]);
      })
    })
  }
  </script>