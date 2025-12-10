<template>
    <div class="w-full flex flex-col gap-2 overflow-auto" style="max-height: calc(100vh - 220px);">
        <div class="sticky top-0 bg-white z-10">
            <div class="flex justify-end gap-1 mb-2">
                <button class="p-1 px-2 bg-primary text-white  flex gap-1 items-center"
                :class="[data.loading?'pointer-event-none bg-slate-400  ':'hover:bg-secondary']" @click="checkedAll(true)">
                    <mdicon name="check" size="16" />
                    Select all
                </button>
                  <button class="p-1 px-2 bg-primary text-white  flex gap-1 items-center"
                :class="[data.loading?'pointer-event-none bg-slate-400  ':'hover:bg-secondary']" @click="checkedAll(false)">
                    <mdicon name="close" size="16" />
                    Unselect all
                </button>
                <button class="p-1 px-2 bg-primary text-white  flex gap-1 items-center"
                :class="[data.loading?'pointer-event-none bg-slate-400  ':'hover:bg-secondary']" @click="loadRBAC">
                    <mdicon name="refresh" size="16" />
                    Reload
                </button>
                <button class="p-1 px-2 bg-primary text-white   flex gap-1 items-center"  
                :class="[data.loading?'pointer-event-none bg-slate-400 ' :'hover:bg-secondary']" @click="saveRBAC">
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
        </div> 
        <template v-if="data.loading">
            <loader kind="skeleton" skeleton-kind="input"/>
            <loader kind="skeleton" skeleton-kind="list"/>
            <loader kind="skeleton" skeleton-kind="list"/>
        </template>
        <div v-else class="relative">
            <div v-for="fc in Object.keys(data.featureCategories)" class="mb-4 border-b border-slate-300">
                <div class="flex gap-[1px] font-semibold border-b-2 bg-[#f4f9ff]">
                    <div class="p-1 grow">{{ fc=="undefined" ? "No Category" : fc }}</div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="data.featureCategories[fc].Grant.All" @click="checkedHeader('All',fc )"/></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="data.featureCategories[fc].Grant.Read" @click="checkedHeader('Read',fc )"/></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="data.featureCategories[fc].Grant.Create" @click="checkedHeader('Create',fc )"/></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="data.featureCategories[fc].Grant.Update" @click="checkedHeader('Update',fc )"/></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="data.featureCategories[fc].Grant.Delete" @click="checkedHeader('Delete',fc )"/></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="data.featureCategories[fc].Grant.Posting" @click="checkedHeader('Posting',fc )"/></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="data.featureCategories[fc].Grant.Special1" @click="checkedHeader('Special1',fc)"/></div>
                    <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="data.featureCategories[fc].Grant.Special2" @click="checkedHeader('Special2',fc )"/></div>
                 </div>
                <div v-for="(f,idx) in data.featureCategories[fc].Details" :key="idx">
                    <div class="flex gap-[1px] hover:bg-slate-100">
                        <div class="grow p-1">{{ f.Feature.Name }}</div>
                        <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.All" @click="checkedDetail('All',fc, idx)"/></div>
                        <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Read" @click="checkedDetail('Read',fc, idx)"/></div>
                        <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Create" @click="checkedDetail('Create',fc, idx)"/></div>
                        <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Update" @click="checkedDetail('Update',fc, idx)"/></div>
                        <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Delete" @click="checkedDetail('Delete',fc, idx)"/></div>
                        <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Posting" @click="checkedDetail('Posting',fc, idx)"/></div>
                        <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Special1" @click="checkedDetail('Special1',fc, idx)"/></div>
                        <div class="p-1 w-[50px] text-center"><input type="checkbox" v-model="f.Grant.Special2" @click="checkedDetail('Special2',fc, idx)"/></div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { inject, onMounted, reactive, resolveDirective } from 'vue';
import { util } from 'suimjs';
import Loader from "@/components/common/Loader.vue";

const props = defineProps({
    role: { type: Object, default: () => { } }
})

const axios = inject("axios")

const data = reactive({
    featureCategories: [],
    loading: false
})

function groupBy(xs) {
    return xs.reduce(function (rv, x) {
        const categoryID = x.Feature.FeatureCategoryID;
        rv[categoryID] = rv[categoryID] || [];
        rv[categoryID].push(x);
        return rv;
    }, {});
}
function createGrant(val = false){
    return {
        All: val,
        Create: val,
        Read: val,
        Update: val,
        Delete: val,
        Posting: val,
        Special1: val,
        Special2: val
    }
}

function getHeaderGrant(details){
    let g = createGrant(true)
    details.forEach(el=>{
        Object.keys(el.Grant).forEach(x =>{
            if(g[x] === true){
                g[x] = el.Grant[x]
            }
        })
    })
    return g

}

function loadRBAC() {
    data.loading = true
    axios.post("admin/feature/find").then(r => {
        const features = r.data.map(el => {
            return {
                AccountID: el.FeatureAccountID,
                Feature: el,
                Grant: createGrant()
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
                        Posting: rf.Posting,
                        Special1: rf.Special1,
                        Special2: rf.Special2,
                    }
                }
            })
            const details =  groupBy(features, "FeatureCategoryID")

            Object.keys(details).forEach(k=>{
                details[k] = 
                 {
                    Details: details[k],
                    Grant: getHeaderGrant(details[k])

                }
            })

            data.featureCategories =details
            data.loading = false;
        })
    }, e => { 
            data.loading = false;
            util.showError(e)
    })
}

function saveRBAC() {
    data.loading = true
    let features = []
    Object.keys(data.featureCategories).forEach(fc => {
        data.featureCategories[fc].Details.forEach(f => {
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
    }, e => util.showError(e)).finally(()=>{
        data.loading = false
    })
}
function checked(obj,key,value){ 
    if(key === 'All'){
        obj.Grant = createGrant(value)
    }else if(value === false){
        obj.Grant.All = false 
    }else{
        let isAllTrue = true 
        Object.keys(obj.Grant).forEach(k => {
         
           if(!['All', key].includes(k)){
             if(!obj.Grant[k]) isAllTrue = false
           }
        });
        obj.Grant.All = isAllTrue
    }

}
function checkedDetail(grant,category, idx){
    const obj = data.featureCategories[category].Details[idx]
    obj.Grant[grant] = !obj.Grant[grant]
    const nv =  obj.Grant[grant]

    checked(obj,grant, nv)

    
    util.nextTickN(2, () => {
        data.featureCategories[category].Grant = getHeaderGrant(data.featureCategories[category].Details)
    })
}
function checkedHeader(grant,category){
    const header = data.featureCategories[category]
    const nv =  !header.Grant[grant]
    
    checked(header, grant, nv)

    data.featureCategories[category].Details.forEach((el,i)=>{
        el.Grant[grant] = !el.Grant[grant]
        checked(el,grant,nv)
    })

}
function checkedAll(val){
    Object.keys(data.featureCategories).forEach(fc => {
        data.featureCategories[fc].Grant = createGrant(val)
        data.featureCategories[fc].Details.forEach(e=>{
            e.Grant = createGrant(val)
        })
    }) 
}
onMounted(() => {
    loadRBAC()
})

</script>