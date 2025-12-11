<template>
    <s-card title="Link User to Profile" class="card border rounded-md">
    <s-form class="min-w-[500px] max-w-[800px] flex flex-col gap-2" 
        buttons-on-bottom
        submit-text="Link" cancel-text="Back" @cancel-form="backToGrid"
        :config="data.linkFormConfig" v-model="data.record"
        @submit-form="linkUser"
        @field-change="handleFieldChange">
    </s-form>
    </s-card>
</template>

<script setup>

import { SCard, SForm, SButton, createFormConfig, formInput, util } from 'suimjs';
import { inject, onMounted, reactive } from 'vue';

const axios = inject("axios")

const props = defineProps({
    profile: {type: Object, default: () => {
        return {
            UserID:'',
            Name:'',
            Email:'',
            LinkToEmail:'',
            LinkToUserID:'',
            LinkToName:''
        }
    }}
})

const data = reactive({
    record: props.profile,
    linkFormConfig: {}
})

const emit = defineEmits({
    cancel: null,
    close: null,
})

function backToGrid () {
    emit("cancel")
}

function handleFieldChange (field, v1, v2, old) {
    if (field=="LinkToEmail") {
        axios.post("/iam/GetUserBy",{FindBy:'_id', FindID:v1}).then(r => {
            if (r.data.error) {
                data.record.LinkToName = ''
                data.record.LinkToUserID = ''
                return    
            }

            data.record.LinkToName = r.data.DisplayName
            data.record.LinkToUserID = r.data._id
        }, e => {
            data.record.LinkToName = ''
        })
    }
}

function linkUser (_, cb1, cb2) {
    axios.post("/karamaster/kara/LinkUser",{
        UserID: data.record.LinkToUserID,
        ProfileID: data.record._id,
    }).then(r => {
        util.showInfo("profile has been linked to user "+data.record.LinkToName)
        cb1()
        emit('close')
    },
    
    e => {
        util.showError(e)
        cb2()
    })
}

function userPayloadBuilder (v) {
    return {
      FindID: v,
      FindBy: 'Email'
    }
  }

onMounted(() => {
    const cfg = createFormConfig("linkForm", false)
    cfg.addSection("Profile", true).addRowAuto(2, 
        new formInput({field:"Name", label:"Name", kind:"string", readOnly:true}),
        new formInput({field:"Email", label:"Email", kind:"string", readOnly:true}))

    cfg.addSection("Link To", true).addRowAuto(2,
        new formInput({field:"LinkToName", label:"Name", required:true, kind:"string", readOnly:true}),
        new formInput({field:"LinkToEmail", label:"Email", kind:"string",
            useList:true, lookupUrl:'/iam/FindUserBy', required:true,
                lookupKey:'_id', lookupLabels:['Email'], lookupSearchs:['Email'],
                lookupPayloadBuilder:userPayloadBuilder}))

    data.linkFormConfig = cfg
})

</script>