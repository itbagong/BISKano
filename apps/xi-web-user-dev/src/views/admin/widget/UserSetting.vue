<template>
    <div class="flex flex-col gap-4">
        <div>
            <h3>Tenant Access</h3>
            <s-list-editor no-gap
                :config="data.gridConfig" v-model="data.tenantUsers" 
                allow-add hide-header hide-select
                @validate-item="validateTenant" @delete-record="deleteUserTenant" @save-record="saveUserTenant"
                ref="tenantUsersCtl">
                <template #item="{item}">
                    <div class="w-[300px]">{{ item.TenantID }}</div>
                    <div class="grow">{{ item.TenantName }}</div>
                </template>
                <template #editor>
                    <div class="flex gap-2">
                        <s-input class="min-w-[300px]" ref="tenantSelector"
                            v-model="data.tenantUser.TenantID" 
                            label="Tenant" 
                            use-list
                            lookup-url="/admin/tenant/find"
                            lookup-key="_id"
                            :lookup-labels="['FID','Name']"
                        />
                    </div>
                </template>
            </s-list-editor>
        </div>

        <div>
            <h3>Change Password</h3>
            <div class="flex gap-2 items-start justify-start">
                <s-input ref="newPass" kind="password" v-model="data.changePassword.newPass" hide-label :rules="rules1" class="w-[200px]"></s-input>
                <s-input ref="confirmPass" kind="password" v-model="data.changePassword.confirmPass" hide-label :rules="rules2" class="w-[200px]"></s-input>
                <s-button class="bg-primary text-white" label="Change" @click="changePassword" :disabled="disableChangePassword"></s-button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { computed, inject, onMounted, reactive, ref } from 'vue';
import { SListEditor, SInput, SButton, util, rules } from 'suimjs';

const props = defineProps({
    user: {type: Object, default: () => {}}
})
const tenantUsersCtl = ref(null)
const tenantSelector = ref(null)
const newPass = ref(null)
const confirmPass = ref(null)
const axios = inject("axios")

const data = reactive({
    gridConfig: {},
    tenantUsers: [],
    tenantUser: {
        TenantID: ""
    },
    changePassword: {
        newPass: "",
        confirmPass: ""
    }
})


const rules1 = [rules.strongPassword(8)]
const rules2 = [(v) => {
    return v!=data.changePassword.newPass ? "new password and confirm should be same" : ""
}]

const disableChangePassword = computed({
    get: () => {
        if (newPass.value==undefined) return false
        if (confirmPass.value==undefined) return false
        return !newPass.value.isValid() || !confirmPass.value.isValid()
    }
})

function changePassword () {
    //if (data.changePassword.newPass != data.changePassword.confirmPass) util.showError('password and confirm should be same')
    axios.post("/admin/change-password",{UserID:props.user._id, Password:data.changePassword.newPass}).then(r=>{
        util.showInfo("Password has been changed successfully")
    }, e => util.showError(e))
}

function validateTenant (record) {
    record.TenantID = data.tenantUser.TenantID
    record.TenantName = tenantSelector.value.value2().split("|")[1].trim(" ")
    tenantUsersCtl.value.setValidateItem(true)
}

function deleteUserTenant (record) {
    if (record._id==undefined || record._id=="") return

    axios.post("/admin/tenantuser/delete", record)
}

function saveUserTenant (record) {
    //alert(record)
    record.UserID = props.user._id
    axios.post("/admin/tenantuser/save", record)
}

onMounted(() => {
    axios.post("/admin/tenantuser/find?UserID="+props.user._id).then(r => {
        data.tenantUsers = r.data
        //tenantUsersCtl.value.setRecord(data.tenantUsers)
    }, e => util.showError(e))
})

</script>