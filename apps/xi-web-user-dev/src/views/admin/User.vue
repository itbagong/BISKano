<template>
    <div class="w-full">
        <data-list class="card" ref="grid" title="User" grid-config="/admin/user/gridconfig"
            form-config="/admin/user/formconfig" grid-read="/admin/user/gets" form-read="/admin/user/get" grid-mode="grid"
            grid-delete="/admin/user/delete" form-insert="/admin/user/save" grid-hide-select form-update="/admin/user/save"
            :init-app-mode="data.appMode" :init-form-mode="data.formMode" :form-fields="['Status', 'Dimension']" :grid-fields="['Dimension']"
            :form-tabs-edit="['General', 'Setting', 'Access']" stay-on-form-after-save @formNewData="newRecord"
            @formEditData="openForm" grid-hide-delete @alterGridConfig="alterGridConfig" @alterFormConfig="alterFormConfig">

            <template #grid_Dimension="{ item }">
                {{  item.Dimension ? item.Dimension.map(x => (x.Key ? x.Key : "()") + "=" + x.Value).join(", ") : "" }}
            </template>

            <template #form_tab_Setting="{ item }">
                <user-setting :user="item" class="mb-5"></user-setting>
            </template>

            <template #form_tab_Access="{ item }">
                <user-grant :user="item" class="mb-5"></user-grant>
            </template>

            <template #form_input_Status="{ item }">
                <div class="flex gap-2 items-start justify-start">
                    <div>
                        <label class="input_label">Status</label>
                        <p>{{ item.Status }}</p>
                    </div>
                    <div class="flex w-full items-center justify-end">
                        <s-button v-if="item.Status == 'Registered'" icon="account-reactivate"
                            class="bg-primary text-white " label="Activate" @click="activateUser(item)"></s-button>
                    </div>
                </div>
            </template>

            <template #form_input_Dimension="{ item }">
                <key-value-editor v-model="item.Dimension"></key-value-editor>
            </template>
        </data-list>
    </div>
</template>

<script setup>
import { nextTick, reactive, ref, inject } from "vue";
import { DataList, SButton, util } from "suimjs";
import UserSetting from "./widget/UserSetting.vue";
import UserGrant from "./widget/UserGrant.vue";
import KeyValueEditor from "@/components/common/KeyValueEditor.vue";
import { layoutStore } from "@/stores/layout";
import { authStore } from "@/stores/auth";

// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
layoutStore().name = "tenant";

const grid = ref(null);

const data = reactive({
    appMode: "grid",
    formMode: "edit",
});

const axios = inject("axios")

function newRecord(record) {
    record._id = "";
    record.Enable = true;

    openForm(record)
}

function activateUser(item) {
    axios.post("/admin/user/activate", item._id).then(r => {
        util.showInfo('User has been activated');
        grid.value.cancelForm()
    }, e => util.showError(e))
}

function alterGridConfig(cfg) {
    cfg.fields = cfg.fields.filter(el => ["WalletAddress", "Enable", "_id"].indexOf(el.field) == -1);
    cfg.setting.sortable = ['LoginID', 'DisplayName'];
    cfg.setting.keywordFields = ['LoginID', 'DisplayName'];
}

function alterFormConfig(cfg) {
    cfg.sectionGroups = cfg.sectionGroups.map(sectionGroup => {
        sectionGroup.sections = sectionGroup.sections.map(section => {
            section.rows.map(row => {
                row.inputs = row.inputs.filter(input => ["Enable"].indexOf(input.field) == -1).map(input => {
                    if (input.field == 'Status') {
                        input.readOnly = true
                        input.items = ["Registered", "Active", "Disable"]
                    }
                    if (input.field == "WalletAddress") {
                        input.hide = true
                    }
                    return input
                })
                return row
            })
            return section
        })
        return sectionGroup
    })
}

function openForm(r) {
    if (r.Dimension == null ) r.Dimension = [];
    
    nextTick(() => {
        nextTick(() => {
            const minLength = 5
            const validChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstyuvwxyz@._-+"
            grid.value.setFormFieldAttr("LoginID", "rules", [
                (v) => {
                    let vLen = 0;
                    let consistsInvalidChar = false;

                    v.split("").forEach((ch) => {
                        vLen++;
                        const validChar = validChars.indexOf(ch) >= 0;
                        if (!validChar) consistsInvalidChar = true;
                    });

                    if (vLen < minLength || consistsInvalidChar)
                        return "minimal length is " + minLength + " and only alphabet and certains whitespace @._-+";
                    return "";
                },
            ]);
        })
    })
}
</script>