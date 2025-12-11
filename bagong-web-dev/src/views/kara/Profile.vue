<template>
    <div>
        <data-list class="card" ref="listControl" grid-mode="grid" form-keep-label title="User Profile" 
            grid-hide-delete
            grid-config="/kara/profile/gridconfig" 
            grid-read="/kara/profile/gets"
            form-config="/kara/profile/formconfig"
            form-read="/kara/profile/get"
            form-insert="/kara/admin/profile/insert"
            form-update="/kara/admin/profile/update"
            stay-on-form-after-save
            :form-tabs-edit="['General','Leave Balance']"
            :grid-hide-new="!profile.canCreate"
            :grid-hide-edit="!profile.canUpdate" 
        >
            <template #form_tab_Leave_Balance="props">
                <data-list no-gap grid-mode="grid" 
                    grid-hide-search grid-hide-sort grid-hide-select
                    grid-config="/kara/leavebalance/gridconfig"
                    :grid-read="`/kara/leavebalance/gets?UserID=${props.item.UserID}`"
                    form-config="/kara/leavebalance/formconfig"
                    form-insert="/kara/leavebalance/insert"
                    form-update="/kara/leavebalance/update"
                    grid-update="/kara/leavebalance/update"
                    grid-delete="/kara/leavebalance/delete"
                    @form-new-data="newRecord"
                >
                </data-list>
            </template>
        </data-list>
    </div>
</template>

<script setup>
</script>

<script setup>
import { reactive, ref, onMounted, watch, inject } from "vue";
import { authStore } from "@/stores/auth.js";
import { layoutStore } from "@/stores/layout.js";
import { DataList, SCard, util, loadFormConfig } from "suimjs";
import UserLocation from "./widget/UserLocation.vue";

layoutStore().name = "tenant";


const FEATUREID = 'UserProfile'
const profile = authStore().getRBAC(FEATUREID)


const axios = inject("axios");
const listControl = ref(null);
const data = reactive({
    record: {
        profileID: null,
    }
});

function newRecord(record) {
    const gridRecord = listControl.value.getFormRecord();
    record.UserID =  gridRecord.UserID;
}

</script>