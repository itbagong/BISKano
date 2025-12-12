<template>
    <div class="w-full flex flex-col gap-2">
        <div class="flex justify-end gap-1">
            <button class="p-1 px-2 bg-primary text-white hover:bg-secondary flex gap-1 items-center" @click="loadRBAC">
                <mdicon name="refresh" size="16" />
                Reload
            </button>
            <button class="p-1 px-2 bg-primary text-white hover:bg-secondary flex gap-1 items-center" @click="saveRBAC">
                <mdicon name="content-save" size="16" />
                Save
            </button>
        </div>
        <div class="flex gap-[1px] mb-2">
            <div class="bg-slate-200 p-1 grow">Name</div>
            <div class="bg-slate-200 p-1 w-[50px] text-center">All</div>
            <div class="bg-slate-200 p-1 w-[50px] text-center">R</div>
            <div class="bg-slate-200 p-1 w-[50px] text-center">C</div>
            <div class="bg-slate-200 p-1 w-[50px] text-center">U</div>
            <div class="bg-slate-200 p-1 w-[50px] text-center">D</div>
            <div class="bg-slate-200 p-1 w-[50px] text-center">P</div>
            <div class="bg-slate-200 p-1 w-[50px] text-center">S1</div>
            <div class="bg-slate-200 p-1 w-[50px] text-center">S2</div>
        </div>
        <div v-for="fc in Object.keys(data.featureCategories)" class="mb-4 border-b border-slate-300">
            <div class="p-1 font-semibold border-b-2">{{ fc=="undefined" ? "No Category" : fc }}</div>
            <div v-for="f in data.featureCategories[fc]">
                <div class="flex gap-[1px] hover:bg-slate-100">
                    <div class="grow p-1">{{ f.Feature.Name }}</div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.All" /></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Read" /></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Create" /></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Update" /></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Delete" /></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Posting" /></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Special1" /></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Special2" /></div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { inject, onMounted, reactive, resolveDirective } from 'vue';
import { util } from 'suimjs';

const props = defineProps({
    role: { type: Object, default: () => { } }
})

const axios = inject("axios")

const data = reactive({
    featureCategories: []
})

function groupBy(xs) {
    return xs.reduce(function (rv, x) {
        const categoryID = x.Feature.FeatureCategoryID;
        rv[categoryID] = rv[categoryID] || [];
        rv[categoryID].push(x);
        return rv;
    }, {});
}

function loadRBAC() {
    axios.post("admin/feature/find").then(r => {
        const features = r.data.map(el => {
            return {
                AccountID: el.FeatureAccountID,
                Feature: el,
                Grant: {
                    All: false,
                    Create: false,
                    Read: false,
                    Update: false,
                    Delete: false,
                    Posting: false,
                    Special1: false,
                    Special2: false,
                }
            }
        })
        //console.log(data.featureCategories)

        axios.post("admin/rolefeature/find?RoleID=" + props.role._id).then(rfs => {
            rfs.data.forEach(rf => {
                const filtered = features.filter(f => f.Feature._id == rf.FeatureID)
                if (filtered.length > 0) {
                    filtered[0].Grant = {
                        All: rf.All,
                        Create: rf.Create,
                        Read: rf.Read,
                        Update: rf.Update,
                        Delete: rf.Delete,
                        Post: rf.Posting,
                        Special1: rf.Special1,
                        Special2: rf.Special2,
                    }
                }
            })

            data.featureCategories = groupBy(features, "FeatureCategoryID")
            console.log(data.featureCategories)
        })
    }, e => util.showError(e))
}

function saveRBAC() {
    let features = []
    Object.keys(data.featureCategories).forEach(fc => {
        data.featureCategories[fc].forEach(f => {
            let checked = false

            const grantKeys = Object.keys(f.Grant)
            for (let kIdx = 0; kIdx < grantKeys.length; kIdx++) {
                const k = grantKeys[kIdx]
                if (f.Grant[k]) {
                    checked = true
                    break
                }
            }    

            if (checked) {
                let g = f.Grant
                g.FeatureID = f.Feature._id
                features.push(g)
            }
        })
    })

    const payload = {RoleID: props.role._id, Features: features}
    axios.post("admin/add-features-to-role", payload).then(r => {
        util.showInfo("feature setting has been saved")
    }, e => util.showError(e))
}

onMounted(() => {
    loadRBAC()
})

</script>