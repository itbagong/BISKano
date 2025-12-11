<template>
    <div class="card w-full p-4 flex flex-col gap-2">
        <div class="flex">
            <h1>Apply</h1>
            <div class="grow">&nbsp;</div>
            <div class="align-right flex gap-1">
                <s-button class="bg-warning" label="Back" icon="keyboard-backspace" @click="back" v-if="!hideBack" />
                <s-button class="bg-primary" icon="file-edit-outline" 
                    tooltip="Show/hide cash" 
                    :class="{'opacity-50':!data.showCash}"
                    @click="data.showCash=!data.showCash" />
                <s-button class="bg-primary" icon="cash" 
                    tooltip="Show/hide adjustment"
                    :class="{'opacity-50':!data.showAdjustment}"
                    @click="data.showAdjustment=!data.showAdjustment" />
                <s-button class="bg-primary" icon="plus" tooltip="Add Adjustment Type" 
                    @click="data.showAdjustmentType=true" />
            </div>
        </div>

        <div 
            v-if="data.showAdjustmentType==false"
            class="flex gap-2 h-[570px]">
            <div v-if="data.hideFromApply">
                <s-button class="bg-primary" icon="eye-outline" tooltip="Hide apply from" @click="data.hideFromApply=false" />
            </div>
            <div class="flex flex-col gap-2 max-w-[450px]" v-if="!data.hideFromApply">
                <div class="flex gap-[1px] items-center justify-start">
                    <template v-if="!hideFromFilter">
                        <s-input label="Source Type" hide-label  v-model="data.sourceFilter.sourceType" use-list :items="['Vendor','Cashbank','Customer','Ledger']" class="w-[150px]" />    
                        <s-input label="Journal ID" hide-label  v-model="data.sourceFilter.sourceJournalID" />    
                    </template>
                    <s-button class="bg-primary" icon="refresh" tooltip="Refresh" @click="getSources" />
                    <s-button class="bg-primary" icon="eye-off" tooltip="Hide apply from" @click="data.hideFromApply=true" />
                </div>

                <div class="flex gap-2 p-2 bg-secondary">
                    <div class="grow">Apply From</div>
                    <div class="w-[120px] text-right">Remaining/Amount</div>
                </div>

                <div class="flex flex-col overflow-auto scroll-auto">
                    <div v-for="(source, index) in data.sources" @click="selectSource(index)"
                        class="flex gap-1 p-2 even:bg-slate-200 cursor-pointer hover:bg-slate-300" 
                        :class="{'border-solid border-primary border-[1px] bg-slate-100':data.selectedIndex==index}"
                    >
                        <div class="grow">
                            <div class="text-[0.8em]">
                                {{ source.Expected }} / 
                                {{ `${source.SourceType} ${source.SourceJournalID}` }} / 
                                Voucher {{ source.VoucherNo }}</div>
                            <div class="text-[0.8em]">{{ source.Account.AccountType }} {{ source.Account.AccountID }}</div>
                            <div class="font-semibold">{{ source.Text }}</div>
                        </div>
                        <div class="text-right w-[120px]">
                            <div class="px-2"
                                :class="{
                                    'bg-slate-300 text-slate-700':source.CalcOutstanding==0, 
                                    'bg-warning text-slate-700':source.CalcOutstanding!=source.Amount && source.CalcOutstanding!=0, 
                                    'bg-[#33AA00] text-white':source.CalcOutstanding==source.Amount
                                }"
                            >{{  util.formatMoney(source.CalcOutstanding) }}<br/>{{  util.formatMoney(source.Amount) }}</div>
                        </div>
                    </div>
                </div>
            </div>

            <div class="grow pl-2 border-l-[1px] border-slate-300 flex flex-col gap-2">
                <div class="flex gap-[1px] justify-end items-center">
                    <template v-if="!hideToFilter">
                        <s-input label="Source Type" hide-label  v-model="data.sourceFilter.keyword" use-list :items="['Cashbank','Vendor','Customer','Ledger']" class="w-[150px]" />    
                        <s-input label="Keyword" hide-label  v-model="data.sourceFilter.keyword" />    
                        <s-input label="Source Account Type" hide-label  v-model="data.sourceFilter.keyword" use-list :items="['Cashbank','Vendor','Customer','Ledger']" class="w-[150px]" />    
                        <s-input label="Source Account" hide-label  v-model="data.sourceFilter.keyword" use-list :items="['Cashbank','Vendor','Customer','Ledger']" class="w-[150px]" /> 
                    </template>
                    <s-button class="bg-primary" tooltip="Refresh" icon="refresh" @click="selectSource(data.selectedIndex)" />
                    <s-button class="bg-primary" tooltip="Confirm" icon="content-save" @click="saveApply" />
                </div>

                <div class="flex gap-2 p-2 bg-secondary">
                    <div class="w-[50px]"><input type="checkbox" @change="checkUncheckAll" /></div>
                    <div class="w-[250px]">Apply to</div>
                    <div class="w-[120px] text-right">Remaining/Amount</div>
                    <div v-if="data.showCash" class="w-[120px] text-right">Apply Amount</div>
                    <div v-if="data.showCash" class="w-[80px] text-right">&nbsp;</div>
                    <div v-if="data.showAdjustment" class="w-[100px]" v-for="adj in data.adjustmentTypes">{{ adj._id }}<br/>{{ adj.Name }}</div>
                </div>

                <div class="flex flex-col overflow-auto scroll-auto">
                    <div class="flex gap-1 p-2 even:bg-slate-200" v-for="dest in data.dests">
                        <div class="w-[50px]"><input type="checkbox" v-model="dest.isSelected" @change="(e) => {
                             const checked = e.target.checked;
                             if (checked) { 
                                fullApply(dest)
                             } else {
                                cancelApply(dest)
                             }
                        }" /></div>
                        <div class="w-[250px]">
                            <div class="text-[0.8em]">{{ dest.Expected }} / 
                                {{ dest.SourceType}} {{ dest.SourceJournalID }} / 
                                Voucher {{ dest.VoucherNo }}</div>
                            <div class="text-[0.8em]">{{ dest.Account.AccountType }} {{ dest.Account.AccountID }}</div>
                            <div class="font-semibold">{{ dest.Text }}</div>
                        </div>
                        <div class="text-right w-[120px]">
                            <div class="px-2"
                                :class="{
                                    'bg-slate-300 text-slate-700':dest.CalcOutstanding==0, 
                                    'bg-warning text-slate-700':dest.CalcOutstanding!=dest.Amount && dest.CalcOutstanding!=0, 
                                    'bg-[#33AA00] text-white':dest.CalcOutstanding==dest.Amount
                                }"
                            >{{  util.formatMoney(dest.CalcOutstanding) }}<br/>{{  util.formatMoney(dest.Amount) }}</div>
                        </div>
                        <s-input v-if="data.showCash" kind="number" hide-label class="w-[120px] ml-[10px]" v-model="dest.Apply"
                            :disabled="(dest.CalcOutstanding==0 || data.selectedSource.CalcOutstanding==0) && !data.selectedSource.Draft"  
                            @change="onApplyAmountChange(dest)"
                        />
                        <div v-if="data.showCash" class="flex gap-[1px] w-[80px]">
                            <s-button class="bg-primary h-[30px]" tooltip="Full apply" icon="check-all"
                                :disabled="(dest.CalcOutstanding==0 || data.selectedSource.CalcOutstanding==0) && !data.selectedSource.Draft"
                                @click="fullApply(dest)"
                            />
                            <s-button class="bg-primary h-[30px]" tooltip="Cancel apply" icon="cancel"
                                v-if="dest.Apply!=0"
                                @click="cancelApply(dest)"
                            />
                        </div>

                        <s-input 
                            v-if="data.showAdjustment"     
                            v-for="adj in data.adjustmentTypes" kind="number" hide-label class="w-[100px]" v-model="dest.Adjs[adj._id]"
                                @change="onAdjustmentAmountChange(dest, adj)" />
                    </div>
                </div>
            </div>
        </div>

        <!-- adjusment type -->
        <div v-if="data.showAdjustmentType==true" class="flex gap-2">
            <s-input class="grow"
                v-model="data.adjustmentTypeAccounts"
                use-list multiple 
                lookup-url="/tenant/ledgeraccount/find"
                lookup-key="_id"
                :lookup-labels="['_id','Name']"
                :lookup-searchs="['_id','Name']"
            />
            <s-button class="bg-primary" icon="content-save" tooltip="Save adjustment type"
                 @click="saveAdjustmentType" />
        </div>
    </div>

    <s-modal 
        ref="dlgFinal"
        title="Save Apply & Adjustment"
        :display="false"
    >
        Following apply and adjusment will be saved:
    </s-modal>
</template>

<script setup>
import { SButton, SInput, SModal, util } from 'suimjs';
import { inject, ref, onMounted, reactive } from 'vue';

const props = defineProps({
    draft: {type: Boolean, default: false},
    hideBack: {type: Boolean, default: false},
    hideFromFilter: {type: Boolean, default: false},
    hideToFilter: {type: Boolean, default: false},
    sourceType: {type: String, default: "CashBank"},
    sourceJournalId: {type: String, default: "66114d0512253a197c38b761"}
})

const emit = defineEmits({
    back: null
})

const dlgFinal = ref(null);
const axios = inject("axios");

const data = reactive({
    sourceFilter: {
        direction: "",
        keyword: "",
        draft: props.draft,
        sourceType: props.sourceType,
        sourceJournalID: props.sourceJournalId
    },
    showFinal: false,
    showAdjustment: true,
    showCash: true,
    adjustmentTypeAccounts: [],
    adjustmentTypes: [],
    showAdjustmentType: false,
    hideFromApply: false,
    fromLoading: false,
    toLoading: false,
    sources: [],
    dests: [],
    applies:[],
    adjustments: [],
    selectedIndex: undefined,
    selectedSource: undefined    
})

function back() {
    emit('back');
    emit('saveApply');
}

function getSources() {
    axios.post("/fico/apply/get-schedules",{
        SourceType: data.sourceFilter.sourceType,
        SourceJournalID: data.sourceFilter.sourceJournalID,
        Draft: data.sourceFilter.draft,
        Direction: data.sourceFilter.direction,
    }).then(r => {
        data.sources = r.data.Schedules.map(el=>{
            el.Draft = data.sourceFilter.draft;
            el.CalcOutstanding = el.Outstanding;
            if (props.draft) el.Amount=0;
            return el;
        });
    }, e => util.showError(e))
}

function fullApply(dest) {
    const absSource = data.selectedSource.CalcOutstanding < 0 ? Math.abs(data.selectedSource.CalcOutstanding) : data.selectedSource.CalcOutstanding;
    const absDest = dest.CalcOutstanding < 0 ?  Math.abs(dest.CalcOutstanding) : dest.CalcOutstanding;

    const amount = props.draft ?
        dest.CalcOutstanding : 
        (absSource >= absDest) ? dest.CalcOutstanding : data.selectedSource.CalcOutstanding;
    makeApply(dest, amount);
}

function cancelApply(dest) {
    makeApply(dest, 0);
}

function makeApply(dest, amount) {
    dest.Apply = amount;
    dest.isSelected = amount !== 0
   
    let apply = data.applies.find(el => {
        return el.From==data.selectedSource._id && el.To==dest._id;
    })

    if (apply==undefined) {
        data.applies.push({
            From: data.selectedSource._id,
            To: dest._id,
            Amount: dest.Apply
        });
    } else {
        apply.Amount = dest.Apply;
    }
    
    recalc(data.selectedSource, dest);
}

function saveApply() {
    if (data.applies.length==0 && data.adjustments.length==0) {
        util.showError('No apply or adjustment has been entered');
        return;
    }

    // todo: show summary of apply


    // prepare cash-apply
    let applies = data.applies.map(el => {
        return {
            Draft: props.draft,
            Source: {RecordID: el.From},
            ApplyTo: {RecordID: el.To},
            Amount: el.Amount
        }
    })

    const adjs = data.adjustments.reduce((res, current) => {
        let adjs = res.filter(el => el.Source.RecordID==current.From && el.ApplyTo.RecordID==current.To);
        let adj = {};

        if (adjs.length==0) {
            adj = {
                Draft: props.draft,
                Source: {RecordID: current.From},
                ApplyTo: {RecordID: current.To},
                Adjustment: []
            }

            res.push(adj);
        } else {
            adj = adjs[0];
        }

        adj.Adjustment.push({
            Account: {AccountType: "LEDGERACCOUNT", AccountID: current.AdjID},
            Amount: current.Amount
        });

        return res;
    }, 
    []);

    applies = applies.concat(adjs);

    axios.post("/fico/apply/save", applies).then(r => {
        emit("saveApply", data.applies, data.adjusments);
        emit("back");
        util.showInfo("apply and adjustment entries has been saved");
    }, 
    e => util.showError(e))
}

function saveAdjustmentType() {
    axios.post("/tenant/ledgeraccount/find",{"Select":["_id","Name"],"Where":{op:"$in",field:"_id",value:data.adjustmentTypeAccounts}}).
    then(r => {
        data.adjustmentTypes = r.data.map(el => {
            return {_id: el._id, Name: el.Name}
        });


        data.dests.forEach(el => {
            calcDestApply(el);
        });
    }, e => util.showError(e))
    data.showAdjustmentType=false;
}

function recalc(source, dest) {
    if (source) {
        const sourceApplyAmt = data.applies.
            filter(el => el.From==source._id).
            reduce((acc, el) => acc + el.Amount, 0);

        if (source.Amount==0) source.CalcOutstanding = source.Amount + sourceApplyAmt
            else source.CalcOutstanding = source.Amount - sourceApplyAmt;
    }

    if (dest) {
        const destApplyAmt = data.applies.
            filter(el => el.To==dest._id).
            reduce((acc, el) => acc + el.Amount, 0);

        const destAdjAmt = data.adjustments.
            filter(el => el.To==dest._id).
            reduce((acc, el) => acc + el.Amount, 0);

        dest.CalcOutstanding = dest.Amount - destApplyAmt - destAdjAmt;
    }
}

function calcDestApply(dest) {
    const source = data.selectedSource;
    const apply = data.applies.find(el => el.From==source._id && el.To==dest._id);
    dest.Apply = apply ? apply.Amount : 0;

    data.adjustmentTypes.forEach(adj => {
        const adjEntry = data.adjustments.find(el => el.From==source._id &&  el.To==dest._id && el.AdjID==adj._id);
        dest.Adjs[adj._id] = adjEntry ? adjEntry.Amount : 0;
    })
}

function onApplyAmountChange(dest) {
    util.nextTickN(1, () => makeApply(dest, dest.Apply));
}

function onAdjustmentAmountChange(dest, adj) {
    util.nextTickN(1, () => makeAdjustment(dest, adj));
}

function makeAdjustment(dest, adj) {
    const amount = dest.Adjs[adj._id];
    
    let adjEntries = data.adjustments.filter(el => {
        const eq = el.From==data.selectedSource._id &&
            el.To==dest._id && 
            el.AdjID==adj._id;

        return eq;
    });

    if (adjEntries.length > 0) {
        adjEntries[0].Amount = amount;
    } else {
        const adjEntry = {
            From: data.selectedSource._id,
            To: dest._id,
            AdjID: adj._id,
            Amount: amount
        }
        data.adjustments.push(adjEntry);
    }
    
    recalc(data.selectedSource, dest);

    //-- over
    if ((dest.CalcOutstanding < 0 && dest.Amount > 0) || (dest.CalcOutstanding > 0 && dest.Amount < 0)) {
        makeApply(dest, dest.Apply + dest.CalcOutstanding);
    }
}

async function selectSource(index) {
    data.selectedIndex = index;
    data.selectedSource = data.sources[index];

    let getScheduleParm = {};
    if (data.selectedSource.Account.AccountType=="LEDGERACCOUNT")
        getScheduleParm = {
            Direction: data.sourceFilter.direction,
            SourceTypeExcept: "CASHBANK",
            ApplyFromType: data.selectedSource.SourceType,
            ApplyFromJournalID: data.selectedSource.SourceJournalID
        }
    else 
        getScheduleParm = {
            Direction: data.sourceFilter.direction,
            SourceTypeExcept: "CASHBANK",
            SubledgerType: data.selectedSource.Account.AccountType,
            SubledgerID: data.selectedSource.Account.AccountID,
            ApplyFromRecordID: data.selectedSource._id
        };

    axios.post("/fico/apply/get-schedules", getScheduleParm).then(async r => {
        // calculate applies and adjustment
        r.data.Applies.forEach(apply => {
            const currentApply = data.applies.find(current => {
                const eq = current.From==apply.Source.RecordID &&
                    current.To==apply.ApplyTo.RecordID;
                return eq;
            })
            if (currentApply==undefined) {
                data.applies.push({
                    From: apply.Source.RecordID,
                    To: apply.ApplyTo.RecordID,
                    Amount: apply.ApplyAmount
                });
            }

            apply.Adjustment.forEach(adj => {
                const currentAdj = data.adjustments.find(current => {
                    const eq = current.From==apply.Source.RecordID &&
                        current.To==apply.ApplyTo.RecordID &&
                        current.AdjID==adj.Account.AccountID;
                    return eq;
                })

                if (currentAdj==undefined) {
                    data.adjustments.push({
                        From: apply.Source.RecordID,
                        To: apply.ApplyTo.RecordID,
                        AdjID: adj.Account.AccountID,
                        Amount: adj.Amount
                    })
                }
            })
        });

        // calc adjustment types
        let promises = [];
        data.adjustmentTypeAccounts = [];
        data.adjustmentTypes = [];
        data.adjustments.forEach(async adj => {
            const key = data.adjustmentTypeAccounts.find(el => el==adj.AdjID);
            if (key==undefined) {
                data.adjustmentTypeAccounts.push(adj.AdjID);
                promises.push(axios.post("/tenant/ledgeraccount/get",[adj.AdjID]).then(coa => {
                    const adts = {
                        _id: adj.AdjID,
                        Name: coa.data.Name
                    };
                    data.adjustmentTypes.push(adts);
                }));
            }
        });
        await Promise.all(promises);

        // calc dest
        data.dests = r.data.Schedules.filter(el => el._id!=data.selectedSource._id).map(el => {
            el.Apply = 0;
            el.Adjs = [];
            recalc(undefined, el);
            calcDestApply(el);
            return el;
        });
    }, e => util.showError(e));
}

function checkUncheckAll(ev) {
  const checked = ev.target.checked;
  data.dests.forEach((dest) => {
    if (checked) {
        fullApply(dest)
    } else {
        cancelApply(dest, 0)
    }
  });
}

onMounted(() => {
    getSources();
})

</script>