<template>
    <div>
        <data-list ref="lineCtl" title="Rule Line" no-gap hide-title
            grid-hide-search grid-hide-sort
            grid-config="/kara/ruleline/gridconfig" form-config="/kara/ruleline/formconfig" :grid-read="`/kara/ruleline/gets?RuleID=${rule._id}`"
            grid-editor grid-hide-select
            grid-delete="/kara/admin/ruleline/delete"
            grid-update="/kara/admin/ruleline/update"
            form-read="/kara/ruleline/get" grid-mode="grid" 
            form-insert="/kara/admin/ruleline/insert" 
            form-update="/kara/admin/ruleline/update"
            :grid-fields="['_id','Days']"
            :form-fields="['Days']"
            :init-app-mode="data.appMode" :init-form-mode="data.formMode" 
            @formNewData="newRecord" @form-edit-data="selectRecord" @pre-save="preSave">
            <template #grid_header_search>
            <div class="w-full">
                Rule lines for {{ rule.Name }}
            </div>
            </template>
            <template #grid__id="{item}">{{ item._id.substr(item._id.length-6,6) }}</template>
            <template #grid_Days="{item}">
                {{  displayDaysTxt(item.Days) }}
            </template>
            <template #form_input_Days="{item}">
                <div class="flex flex-col gap-2">
                    <div class="input_label">Days</div>
                    <div class="flex gap-4">
                        <div v-for="(dayName, dayIndex) in ['Sun','Mon','Tue','Wed','Thu','Fri','Sat']">
                            <input type="checkbox" v-model="data.lineDays[dayIndex]"> {{  dayName  }}
                        </div>
                    </div>
                </div>
            </template>
        </data-list>
    </div>
</template>

<script setup>
import { reactive } from 'vue';
import { DataList } from 'suimjs';

const props = defineProps({
    rule: {type: Object, default: () => {}}
})

const data = reactive({
    appMode: "grid",
    formMode: "edit",
    lineDays: []
})

function displayDaysTxt (days) {
    if (days.length==0) return 'All'

    const daysTxt = ['Su','Mo','Tu','We','Th','Fr','St']
    return days.map(el => daysTxt[el]).join(", ")
}

function newRecord (record) {
    record.RuleID = props.rule._id;
    record.PersonPerBlock = 0;
    record.MinimumHour = 0;
}

function selectRecord (record) {
    data.lineDays = [false, false, false, false, false, false, false]
    if (record.Days) {
        record.Days.forEach(element => {
            data.lineDays[element] = true
        })
    }
}

function preSave (record) {
    const days = []
    data.lineDays.forEach((element, index) => {
        if (element==true) {
            days.push(index)
        }
    })
    record.Days = days
}

</script>