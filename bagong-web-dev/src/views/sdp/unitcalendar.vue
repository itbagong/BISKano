<template>
  <div class="w-full inline-block">
    <s-card
      class="bg-white card w-[88vw] md:w-[90vw] xl:w-[94vw]"
      hide-footer
      title="Unit Calendar"
    >
      <div class="pt-2">
        <s-form
          v-if="data.formCfg && data.formCfg.setting && data.appMode == 'form'"
          ref="formCtl"
          v-model="data.record"
          :config="data.formCfg"
          :mode="data.formMode"
          class="pt-2"
          :tabs="['General', 'Line']"
          @submitForm="save"
          @cancelForm="cancelForm"
          @fieldChange="handleFormFieldChange"
          :form-fields="['Dimension']"
          :hide-submit="!profile.canUpdate"
        >
          <template #input_SORefNo="{ item, config }">
            <!-- <s-input
							:field="config.field"
							:kind="config.kind"
							:label="config.label"
							@change="handleFormFieldChange"
							:disabled="
								config.readOnly ||
								(data.formMode == 'edit' && config.readOnlyOnEdit)
							"
							:caption="config.caption"
							:hint="config.hint"
							:multi-row="config.multiRow"
							:use-list="config.useList"
							:items="config.items"
							:rules="config.rules"
							:required="config.required"
							:read-only="config.readOnly"
							:lookup-url="config.lookupUrl"
							:lookup-key="config.lookupKey"
							:allow-add="config.allowAdd"
							:lookup-format1="config.lookupFormat1"
							:lookup-format2="config.lookupFormat2"
							:lookup-payload-builder="
								(search) =>
									lookupPayloadBuilder(search, config, item[config.field])
							"
							:decimal="config.decimal"
							:date-format="config.dateFormat"
							:multiple="config.multiple"
							:keep-label="keepLabel"
							:lookup-labels="config.lookupLabels"
							:lookup-searchs="
								config.lookupSearchs && config.lookupSearchs.length == 0
									? config.lookupLabels
									: config.lookupSearchs
							"
							v-model="item[config.field]"
							:class="{checkboxOffset: config.kind == 'checkbox'}"
							ref="inputs"
						>
						</s-input> -->
            <s-input
							keep-label
              v-model="item.SORefNo"
							:read-only="data.formMode == 'edit' && data.iSOfilled"
							:field="config.field"
							:label="config.label"
							:caption="config.caption"
              use-list
							:required="config.required"
							lookup-key="_id"
							:lookup-labels="['SalesOrderNo', 'Name']"
							:lookupSearchs="['_id', 'Name', 'SalesOrderNo']"
							lookup-url="/sdp/salesorder/find"
              @change="handleFormFieldChange"
              ref="inputs"
            >
            </s-input>
          </template>
          <template #tab_Line="{ item }">
            <unit-calendar-line
              :IsSOref="item.SORefNo != undefined"
              grid-config="/sdp/unitcalendar/line/gridconfig"
              v-model="item.Lines"
              :assets="data.Assets"
            >
            </unit-calendar-line>
          </template>
          <template #input_Dimension="{ item, config }">
            <DimensionEditorVertical
              v-model="item.Dimension"
              :read-only="config.readOnly || config.readOnlyOnEdit"
            />
          </template>
        </s-form>

        <div v-show="data.appMode == 'calendar'" class="">
          <div class="flex gap-2 justify-center items-center header mb-6">
            <s-input
              ref="refAsset"
              v-model="data.search.Asset"
              lookup-key="_id"
              label="Asset"
              class="w-full"
              multiple
              use-list
              :lookup-url="`/tenant/asset/find?GroupID=UNT`"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              @change="refreshData"
            ></s-input>
            <s-input
              ref="refSite"
              v-model="data.search.Site"
              lookup-key="_id"
              label="Site"
              class="w-full"
              multiple
              use-list
              :lookup-url="`/bagong/sitesetup/find`"
              :lookup-labels="['Name']"
              :lookup-searchs="['_id', 'Name']"
              @change="refreshData"
            ></s-input>
            <s-input
              kind="date"
              label="Date From"
              v-model="data.search.DateFrom"
              @change="refreshData"
            ></s-input>
            <s-input
              kind="date"
              label="Date To"
              v-model="data.search.DateTo"
              @change="refreshData"
            ></s-input>
            <div class="flex gap-[1px] header_button">
              <s-button
                icon="refresh"
                class="btn_primary refresh_btn"
                tooltip="refresh"
                @click="refreshData"
              />
              <s-button
                icon="plus"
                class="btn_primary new_btn"
                tooltip="add new"
                @click="NewRecord"
                :disabled="!profile.canCreate"
              />
            </div>
          </div>

          <div class="block">
            <div class="relative w-full flex flex-row">
              <!-- <div class="h-full">
								<table class="w-42 text-sm text-center">
									<thead class="text-xs uppercase">
										<tr>
											<th class="h-[36px]"></th>
										</tr>
										<tr>
											<th class="!h-[28px]"></th>
										</tr>
										<tr>
											<th class="!h-[28px]"></th>
										</tr>
									</thead>
									<tbody v-for="site in data.Projects">
										<tr class="text-left">
											<td>
												<s-button
													icon="ChevronDown"
													class="w-42 bg-slate-100 !h-[28px] whitespace-nowrap text-ellipsis overflow-x-hidden"
													:label="site.Label"
												/>
											</td>
										</tr>
										<tr
											v-if="site.visible"
											v-for="line in site.Lines"
											class="text-left"
										>
											<td
												class="w-48 pl-4 text-sm !h-[30px] whitespace-nowrap text-ellipsis overflow-x-hidden cursor-default flex flex-row items-center"
												@click="() => EditRecord(line)"
											>
												<mdicon name="pencil" width="12" />- {{ line.asset }} (
												{{ line.Qty }} )
											</td>
										</tr>
									</tbody>
								</table>
							</div> -->
              <div
                ref="tablescroll"
                class="max-h-[40rem] overflow-auto"
                @scroll="
                  (event) => {
                    if (
                      event.target.scrollLeft + event.target.clientWidth >=
                      event.target.scrollWidth
                    ) {
                      data.dateScroll.DateToMonth += 3;
                    }

                    if (
                      event.target.scrollTop + event.target.clientHeight >=
                      event.target.scrollHeight
                    ) {
                      data.Skip += 1;
                      Projects();
                    }
                  }
                "
              >
                <table
                  class="text-sm text-center text-gray-500 dark:text-gray-400 table-auto tbl"
                >
                  <thead class="text-xs text-gray-800 uppercase sticky-row">
                    <tr>
                      <th class="border-0 sticky-col"></th>
                      <th
                        v-for="month in Object.keys(DateArr)"
                        :colspan="DateArr[month].length"
                        class="!h-[36px]"
                      >
                        {{ month }}
                      </th>
                    </tr>
                    <tr>
                      <th class="border-0 sticky-col"></th>
                      <th
                        v-for="dates in Object.keys(DateArr).flatMap((key) =>
                          DateArr[key].map((date) => ({
                            date: date.getDate(),
                            weekend: date.getDay() === 6 || date.getDay() === 0,
                          }))
                        )"
                        class="border-x border-solid p-1 border-black !h-[28px]"
                        :class="dates.weekend && 'bg-red-200'"
                      >
                        {{ dates.date }}
                      </th>
                    </tr>
                    <tr>
                      <th class="sticky-col"></th>
                      <th
                        v-for="dates in Object.keys(DateArr).flatMap((key) =>
                          DateArr[key].map((date) => ({
                            day: moment(date).format('dd'),
                            weekend: date.getDay() === 6 || date.getDay() === 0,
                          }))
                        )"
                        class="border-x border-solid p-1.5 border-black !h-[28px]"
                        :class="dates.weekend && 'bg-red-200'"
                      >
                        {{ dates.day }}
                      </th>
                    </tr>
                  </thead>
                  <tbody v-for="site in data.Projects" class="tbl-tbody">
                    <!-- <tr class="text-left">
											<td>
												<s-button
													icon="ChevronDown"
													class="w-42 bg-slate-100 !h-[28px] whitespace-nowrap text-ellipsis overflow-x-hidden"
													:label="site.Label"
												/>
											</td>
										</tr>
										<tr
											v-if="site.visible"
											v-for="line in site.Lines"
											class="text-left"
										>
											<td
												class="w-48 pl-4 text-sm !h-[30px] whitespace-nowrap text-ellipsis overflow-x-hidden cursor-default flex flex-row items-center"
												@click="() => EditRecord(line)"
											>
												<mdicon name="pencil" width="12" />- {{ line.asset }} (
												{{ line.Qty }} )
											</td>
										</tr> -->
                    <tr>
                      <td class="sticky-col">
                        <s-button
                          icon="ChevronDown"
                          class="w-42 bg-slate-100 !h-[28px] whitespace-nowrap text-ellipsis overflow-x-hidden text-black"
                          :label="site.Label"
                        />
                      </td>
                      <td
                        v-for="dates in Object.keys(DateArr).flatMap((key) =>
                          DateArr[key].map((date) => {
                            const StartPeriodMonth = new Date(
                              site.StartPeriodMonth
                            );
                            StartPeriodMonth.setHours(0, 0, 0, 0);

                            const EndPeriodMonth = new Date(
                              site.EndPeriodMonth
                            );
                            EndPeriodMonth.setHours(0, 0, 0, 0);
                            const now = new Date();
                            now.setHours(0, 0, 0, 0);

                            return {
                              show:
                                StartPeriodMonth <= date &&
                                date <= EndPeriodMonth,
                              progress: date <= now,
                              dates,
                              StartPeriodMonth,
                              EndPeriodMonth,
                            };
                          })
                        )"
                        class="!h-[28px]"
                        :class="
                          !dates.show && `border-x border-solid border-black `
                        "
                      >
                        <div
                          v-if="dates.show"
                          class="z-50 mx-[-1px] !h-[28px] border-x-[-1px]"
                          :class="
                            dates.progress ? 'bg-purple-700' : 'bg-purple-400'
                          "
                        />
                      </td>
                    </tr>
                    <tr v-if="site.visible" v-for="line in site.Lines">
                      <td
                        class="w-48 pl-4 text-sm !h-[30px] whitespace-nowrap text-ellipsis overflow-x-hidden cursor-pointer hover:bg-slate-200 flex flex-row items-center text-black sticky-col"
                        @click="() => EditRecord(line)"
                      >
                        <mdicon name="pencil" width="12" />- {{ line.asset }} (
                        {{ line.Qty }} )
                      </td>
                      <td
                        v-for="dates in Object.keys(DateArr).flatMap((key) =>
                          DateArr[key].map((date) => {
                            const StartPeriodMonth = new Date(
                              line.StartPeriodMonth
                            );
                            StartPeriodMonth.setHours(0, 0, 0, 0);

                            const EndPeriodMonth = new Date(
                              line.EndPeriodMonth
                            );
                            EndPeriodMonth.setHours(0, 0, 0, 0);
                            const now = new Date();

                            return {
                              show:
                                StartPeriodMonth <= date &&
                                date <= EndPeriodMonth,
                              progress: date <= now,
                              dates,
                              StartPeriodMonth,
                              EndPeriodMonth,
                            };
                          })
                        )"
                        class="!h-[28px]"
                        :class="
                          !dates.show && `border-x border-solid border-black `
                        "
                      >
                        <div
                          v-if="dates.show"
                          class="z-50 mx-[-1px] !h-[28px] border-x-[-1px]"
                          :class="
                            dates.progress ? 'bg-purple-700' : 'bg-purple-400'
                          "
                        />
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
          <div v-if="data.isLoading" class="loading">
            loading data from server ...
          </div>
        </div>
      </div>
    </s-card>
  </div>
</template>

<script setup>
import { SCard, SButton, util, SForm, loadFormConfig, SInput } from "suimjs";
import { layoutStore } from "../../stores/layout.js";
import {
  computed,
  inject,
  watch,
  reactive,
  ref,
  onMounted,
  onBeforeMount,
} from "vue";
import { useRoute } from "vue-router";
import UnitCalendarLine from "./widget/UnitCalendar_Line.vue";
import DimensionEditorVertical from "@/components/common/DimensionEditorVertical.vue";
import moment from "moment";
import "moment/locale/id";
import { authStore } from "@/stores/auth.js";

const formCtl = ref(SForm);

layoutStore().name = "tenant";

const featureID = "UnitCalendar";
// authStore().hasAccess({AccessType:'Role', AccessID:'Administrators'})
// authStore().hasAccess({AccessType:'Feature', AccessID:'LedgerJournal'})
const profile = authStore().getRBAC(featureID);

const axios = inject("axios");

const data = reactive({
  title: null,
  formCfg: undefined,
  formMode: "edit",
  record: {},
  records: [],
  Projects: [],
  appMode: "calendar",
  Assets: [],
  isLoading: false,
  Skip: 0,
  search: {
    Asset: [],
    Site: [],
    DateFrom: null,
    DateTo: null,
  },
  dateScroll: {
    DateFromMonth: 3,
    DateToMonth: 3,
  },
  Sites: [],
	iSOfilled: false,
});

const route = useRoute();
const tablescroll = (ref < HTMLDivElement) | (null > null);

watch(
  () => route.query.objname,
  () => {
    util.nextTickN(2, () => {});
  }
);

watch(
  () => route.query.title,
  (nv) => {
    data.title = nv;
  }
);

const DateArr = computed({
  get() {
    let after = new Date(
      new Date().setMonth(new Date().getMonth() + data.dateScroll.DateToMonth)
    );
    let before = new Date(
      new Date().setMonth(new Date().getMonth() - data.dateScroll.DateFromMonth)
    );
    if (
      data.search.DateFrom !== null &&
      data.search.DateFrom !== "" &&
      data.search.DateFrom !== "Invalid date"
    ) {
      before = new Date(data.search.DateFrom);
    }

    if (
      data.search.DateTo !== null &&
      data.search.DateTo !== "" &&
      data.search.DateTo !== "Invalid date"
    ) {
      after = new Date(data.search.DateTo);
    }

    let days = {};
    let cache = undefined;
    for (var d = before; d <= after; d.setDate(d.getDate() + 1)) {
      const dateloop = new Date(d);
      dateloop.setHours(0, 0, 0, 0);
      const monthstring = moment(dateloop).format("MMMM YYYY");
      if (cache != monthstring) {
        cache = monthstring;
        days[monthstring] = [];
      }
      days[monthstring].push(dateloop);
    }

    return days;
  },
});

function NewRecord() {
  data.record = {
    SORefNo: undefined,
    _id: undefined,
    StartDate: undefined,
    Duration: undefined,
    Customer: undefined,
    AssetUnitID: undefined,
    EndDate: undefined,
    Remark: "",
  };
  data.formMode = "new";
  data.appMode = "form";
}

async function EditRecord(line) {
  data.record = line.record;
	
  data.Assets = await GetAsset([line.line.AssetUnitID]);
  data.record.Lines = [
    {
      Index: line.line.index,
      Duration: line.line.Duration,
      StartDate: line.line.StartDate,
      AssetUnitID: [line.line.AssetUnitID],
      IsItem: false,
      EndDate: line.line.EndDate,
      Uom: line.line.Uom,
      Descriptions: line.line.Descriptions,
      maxQty: 1,
      Qty: 1,
    },
  ];
  if (data.record?.Customer) {
    data.record.CustomerName = (await getDataCustomer(data.record.Customer)).Name;
  }

  data.formMode = "edit";
  data.appMode = "form";
	if (data.record.SORefNo) {
		data.iSOfilled = true
	}
}

async function Init() {
  data.Skip = 0;
  await Projects();
  loadFormConfig(axios, "/sdp/unitcalendar/formconfig").then(
    (r) => {
      data.formCfg = r;
    },
    (e) => util.showError(e)
  );
}

// async function GetProjectWithSites(_idsite) {
// 	const records = await GetRecords(_idsite);

// 	let Assets = await GetAssets(
// 		records.flatMap((record) =>
// 			record.Lines.filter((line) => line.AssetUnitID != "").map(
// 				(line) => line.AssetUnitID
// 			)
// 		)
// 	);

// 	if (data.search.Asset.length > 0) {
// 		Assets = Assets.filter((x) => data.search.Asset.includes(x._id));
// 	}
// 	data.Projects = data.Projects.map((record) => {
// 		if (_idsite == record._id) {
// 			return {
// 				...record,
// 				Lines: records
// 					.filter((record) => {
// 						let Site = (record.Dimension.find((dm) => dm.Key === "Site") ?? {})
// 							.Value;
// 						if (Site == undefined) {
// 							return false;
// 						}

// 						if (_idsite != Site) {
// 							return false;
// 						}

// 						return true;
// 					})
// 					.flatMap((record) => {
// 						return record.Lines.filter((line) => {
// 							if (line.AssetUnitID == null) return false;
// 							const asset = Assets.find(
// 								(asset) => line.AssetUnitID == asset._id
// 							);
// 							if (!asset) return false;
// 							return true;
// 						}).map((line) => {
// 							const asset = Assets.find(
// 								(asset) => line.AssetUnitID == asset._id
// 							);
// 							if (asset) {
// 								return {
// 									line,
// 									record,
// 									index: line.Index,
// 									StartPeriodMonth: line.StartDate,
// 									EndPeriodMonth: line.EndDate,
// 									asset: asset.Name,
// 									Qty: line.Qty,
// 								};
// 							}
// 						});
// 					}),
// 			};
// 		}
// 		return record;
// 	});
// }

async function Projects() {
  try {
    if (data.isLoading) {
      return;
    }
    data.isLoading = true;
    const records = await GetRecords();
    const siterecords = records
      .flatMap(
        (record) =>
          (record.Dimension.find((dm) => dm.Key === "Site") ?? {}).Value
      )
      .filter((record) => record != "" || record != undefined);

    const siteucs = await GetSiteUCs(
      siterecords.filter(
        (record, index) => siterecords.indexOf(record) == index
      )
    );

    let Sites = await GetSiteDimension(
      siteucs.data
        .flatMap(
          (record) =>
            (record.Dimension.find((dm) => dm.Key === "Site") ?? {}).Value
        )
        .filter((record) => record != "" || record != undefined)
    );

    let Assets = await GetAssets(
      records.flatMap((record) =>
        record.Lines.filter((line) => line.AssetUnitID != "").map(
          (line) => line.AssetUnitID
        )
      )
    );

    if (data.search.Asset.length > 0) {
      Assets = Assets.filter((x) => data.search.Asset.includes(x._id));
    }

    if (data.Projects.length > 0) {
      for (const site of Sites) {
        const project = data.Projects.find(
          (project) => site._id == project._id
        );

        if (project) {
          data.Projects = data.Projects.map((project) => {
            return {
              ...project,
              Lines: [
                ...project.Lines,
                ...records
                  .filter((record) => {
                    let Site = (
                      record.Dimension.find((dm) => dm.Key === "Site") ?? {}
                    ).Value;
                    if (Site == undefined) {
                      return false;
                    }

                    if (project._id != Site) {
                      return false;
                    }

                    return true;
                  })
                  .flatMap((record) => {
                    return record.Lines.filter((line) => {
                      if (line.AssetUnitID == null) return false;
                      const asset = Assets.find(
                        (asset) => line.AssetUnitID == asset._id
                      );
                      if (!asset) return false;
                      return true;
                    }).map((line) => {
                      const asset = Assets.find(
                        (asset) => line.AssetUnitID == asset._id
                      );
                      if (asset) {
                        return {
                          line,
                          record,
                          index: line.Index,
                          StartPeriodMonth: line.StartDate,
                          EndPeriodMonth: line.EndDate,
                          asset: asset.Name,
                          Qty: line.Qty,
                        };
                      }
                    });
                  }),
              ],
            };
          });
        } else {
          const siteuc = siteucs.data.find(
            (record) =>
              (
                record.Dimension.find(
                  (dm) => dm.Key === "Site" && dm.Value === site._id
                ) ?? {}
              ).Value
          );

          data.Projects = [
            ...data.Projects,
            {
              ...site,
              visible: true,
              StartPeriodMonth: siteuc.StartDate,
              EndPeriodMonth: siteuc.EndDate,
              Lines: records
                .filter((record) => {
                  let Site = (
                    record.Dimension.find((dm) => dm.Key === "Site") ?? {}
                  ).Value;
                  if (Site == undefined) {
                    return false;
                  }

                  if (site._id != Site) {
                    return false;
                  }

                  return true;
                })
                .flatMap((record) => {
                  return record.Lines.filter((line) => {
                    if (line.AssetUnitID == null) return false;
                    const asset = Assets.find(
                      (asset) => line.AssetUnitID == asset._id
                    );
                    if (!asset) return false;
                    return true;
                  }).map((line) => {
                    const asset = Assets.find(
                      (asset) => line.AssetUnitID == asset._id
                    );
                    if (asset) {
                      return {
                        line,
                        record,
                        index: line.Index,
                        StartPeriodMonth: line.StartDate,
                        EndPeriodMonth: line.EndDate,
                        asset: asset.Name,
                        Qty: line.Qty,
                      };
                    }
                  });
                }),
            },
          ];
        }
      }
    } else {
      let Projects = Sites.map((site) => {
        const siteuc = siteucs.data.find(
          (record) =>
            (
              record.Dimension.find(
                (dm) => dm.Key === "Site" && dm.Value === site._id
              ) ?? {}
            ).Value
        );

        return {
          ...site,
          visible: true,
          StartPeriodMonth: siteuc.StartDate,
          EndPeriodMonth: siteuc.EndDate,
          Lines: records
            .filter((record) => {
              let Site = (
                record.Dimension.find((dm) => dm.Key === "Site") ?? {}
              ).Value;
              if (Site == undefined) {
                return false;
              }

              if (site._id != Site) {
                return false;
              }

              return true;
            })
            .flatMap((record) => {
              return record.Lines.filter((line) => {
                if (line.AssetUnitID == null) return false;
                const asset = Assets.find(
                  (asset) => line.AssetUnitID == asset._id
                );
                if (!asset) return false;
                return true;
              }).map((line) => {
                const asset = Assets.find(
                  (asset) => line.AssetUnitID == asset._id
                );
                if (asset) {
                  return {
                    line,
                    record,
                    index: line.Index,
                    StartPeriodMonth: line.StartDate,
                    EndPeriodMonth: line.EndDate,
                    asset: asset.Name,
                    Qty: line.Qty,
                  };
                }
              });
            }),
        };
      });

      // if (data.search.Site.length > 0) {
      // 	Projects = Projects.filter((x) => data.search.Site.includes(x._id));
      // }
      data.Projects = [...data.Projects, ...Projects];
    }

    data.isLoading = false;
  } catch (error) {
    data.isLoading = false;
    util.showError(error);
  }
}

function lookupPayloadBuilder(search, config, value) {
  const qp = {};
  if (search != "") data.filterTxt = search;
  qp.Take = 20;
  qp.Sort = [config.lookupLabels[0]];
  qp.Select = [...config.lookupLabels];
  let idInSelect = false;
  const selectedFields = config.lookupLabels.map((x) => {
    if (x == config.lookupKey) {
      idInSelect = true;
    }
    return x;
  });
  if (!idInSelect) {
    selectedFields.push(config.lookupKey);
  }
  qp.Select = selectedFields;

  qp.Select.push("Status");

  //setting search
  if (search.length > 0 && config.lookupSearchs.length > 0) {
    if (config.lookupSearchs.length == 1)
      qp.Where = {
        Field: config.lookupSearchs[0],
        Op: "$contains",
        Value: [search],
      };
    else
      qp.Where = {
        Op: "$or",
        items: config.lookupSearchs.map((el) => {
          return { Field: el, Op: "$contains", Value: [search] };
        }),
      };
  }

  if (config.multiple && value && value.length > 0 && qp.Where != undefined) {
    const whereExisting =
      value.length == 1
        ? { Op: "$eq", Field: config.lookupKey, Value: value[0] }
        : {
            Op: "$or",
            items: value.map((el) => {
              return { Field: config.lookupKey, Op: "$eq", Value: el };
            }),
          };

    qp.Where = { Op: "$or", items: [qp.Where, whereExisting] };
  }

  const items = [
    { Op: "$ne", Field: `Dimension`, Value: [] },
    { Op: "$eq", Field: `Status`, Value: "POSTED" },
  ];
  if (qp.Where != undefined) {
    items.push(qp.Where);
  }

  qp.Where = {
    Op: "$and",
    items: items,
  };

  return qp;
}

function cancelForm() {
  data.appMode = "calendar";
}

async function save(saveData, cbOK, cbFalse) {
  const payload = {...saveData}
  try {
    const saveEndPoint =
      data.formMode == "new"
        ? "/sdp/unitcalendar/insert"
        : "/sdp/unitcalendar/update";

    if (saveEndPoint == "") {
      data.appMode = "calendar";
      util.showInfo("data has been saved");
      cbOK();
      return;
    }

    if (payload.Lines.length <= 0) {
      throw Error("Lines not null");
    }

    for (const line of payload.Lines) {
      if (line.maxQty - line.Qty < 0) {
        throw Error("Qty greater than Max Qty");
      }
    }

    const Lines = [];
    for (const line of payload.Lines) {
      for (const asset of line.AssetUnitID) {
        Lines.push({
          ...line,
          AssetUnitID: asset,
          Qty: line.IsItem ? 1 : line.Qty,
        });
      }
    }
    payload.Lines = Lines;

    if (!payload.SORefNo) {
      payload.Customer = payload.CustomerName;
    }

    await axios.post(saveEndPoint, payload);
    data.Projects = [];
    data.Skip = 0;
    await Projects();

    data.appMode = "calendar";

    util.showInfo("data has been saved");
    cbOK();
  } catch (error) {
    util.showError(error);
    cbFalse();
  }
}

async function handleFormFieldChange(name, v1, v2, old) {
  if (name == "SORefNo") {
    const [mp, ucs] = await Promise.all([
      GetSO(v1),
      FindUnitCalendarsSORefNo(v1),
    ]);

    const UCbatchLines = [];

    for (const uc of ucs) {
      for (const line of uc.Lines) {
        if (UCbatchLines.length <= 0) {
          UCbatchLines.push(line);
          continue;
        }

        const UCbatchLine = UCbatchLines.find(
          (UCbatchLine) => UCbatchLine.Index === line.Index
        );

        if (UCbatchLine) {
          UCbatchLines.map((UCbatchLine) => {
            if (UCbatchLine.Index === line.Index) {
              return {
                ...UCbatchLine,
                AssetUnitID: [...UCbatchLine.AssetUnitID, ...line.AssetUnitID],
                Qty: UCbatchLine.Qty + line.Qty,
              };
            }
          });
        } else {
          UCbatchLines.push(line);
        }
      }
    }

    data.record.StartDate = mp.SalesOrderDate;
    data.record.CustomerName = (await getDataCustomer(mp.CustomerID)).Name;
    data.record.Customer = mp.CustomerID;

    const variants = await FindSpecVariants(
      mp.Lines.flatMap((line) => line.Spesifications)
    );
    data.record.Lines = mp.Lines.map((line, index) => {
      const ucline = UCbatchLines.find(
        (UCbatchLine) => UCbatchLine.Index == index
      );

      const datestart = new Date(line.StartDate);
      const dateend = new Date(line.StartDate);

      switch (String(line.UoM).toLowerCase()) {
        case "days":
          dateend.setDate(datestart.getDate() + line.ContractPeriod);
          break;
        case "day":
          dateend.setDate(datestart.getDate() + line.ContractPeriod);
          break;
        case "month":
          dateend.setMonth(datestart.getMonth() + line.ContractPeriod);
          break;
        case "months":
          dateend.setMonth(datestart.getMonth() + line.ContractPeriod);
          break;

        default:
          break;
      }

      let variant = "";
      if (line.Spesifications) {
        variant = line.Spesifications.map(
          (spec) =>
            (variants.find((_variant) => _variant._id == spec) ?? {}).Name
        ).join(" ");
      }

      return {
        Index: index,
        Duration: line.ContractPeriod,
        StartDate: line.StartDate,
        AssetUnitID: line.Asset != "" ? [line.Asset] : [],
        IsItem: Boolean(line.Item),
        EndDate: moment(dateend).local().format("YYYY-MM-DDTHH:mm:ssZ"),
        Uom: line.UoM,
        Descriptions: line.Description + " " + variant,
        maxQty: ucline ? line.Qty - ucline.Qty : line.Qty,
        Qty: !Boolean(line.Item) ? line.Qty : 0,
      };
    }).filter((line) => {
      if (!line.IsItem && line.Qty <= 0) {
        return false;
      }

      if (line.maxQty <= 0) {
        return false;
      }

      return line;
    });

    data.record.Dimension = mp.Dimension;
    data.Assets = await GetAsset(
      mp.Lines.filter((line) => line.Asset != "").map((line) => line.Asset)
    );

    if (
      mp.Dimension.length <= 0 ||
      (mp.Dimension.find((dm) => dm.Key == "Site") ?? {}).Value == ""
    ) {
      util.showError("Dimension site not null");
    }
  }
}

async function GetAssets(_ids) {
  try {
    return (
      await axios.post(`/tenant/asset/find`, {
        Select: ["_id", "Name"],
        Where: {
          OP: "$in",
          Field: "_id",
          Value: _ids,
        },
      })
    ).data;
  } catch (error) {
    util.showError(error);
  }
  return undefined;
}

async function GetRecords() {
  let skip = 0;
  if (data.Skip * 25 != 0) {
    skip = data.Skip * 25 + 1;
  }
  try {
    const payload = {
      Take: 25,
      Skip: skip,
      Sort: ["Site"],
    };

    if (
      (data.search.Site || []).length > 0 ||
      (data.search.Asset || []).length > 0
    ) {
      payload.Where = {
        Op: "$and",
        Items: [],
      };
    }

    if ((data.search.Site || []).length > 0) {
      payload.Where.Items = [
        ...payload.Where.Items,
        {
          Op: "$eq",
          field: "Dimension.Key",
          value: "Site",
        },
        {
          Op: "$in",
          field: "Dimension.Value",
          value: data.search.Site,
        },
      ];
    }

    if ((data.search.Asset || []).length > 0) {
      payload.Where.Items = [
        ...payload.Where.Items,
        {
          Op: "$in",
          field: "Lines.AssetUnitID",
          value: data.search.Asset,
        },
      ];
    }

    return (await axios.post(`/sdp/unitcalendar/find`, payload)).data;
  } catch (error) {
    util.showError(error);
  }
}

async function GetSiteUCs(site) {
  try {
    let payload = {};
    let filters = [];
    // if (
    // 	data.search.DateFrom !== null &&
    // 	data.search.DateFrom !== "" &&
    // 	data.search.DateFrom !== "Invalid date"
    // ) {
    // 	filters.push({
    // 		Field: "StartDate",
    // 		Op: "$gte",
    // 		Value: moment(data.search.DateFrom)
    // 			.utc()
    // 			.format("YYYY-MM-DDT00:mm:00Z"),
    // 	});
    // }

    // if (
    // 	data.search.DateTo !== null &&
    // 	data.search.DateTo !== "" &&
    // 	data.search.DateTo !== "Invalid date"
    // ) {
    // 	filters.push({
    // 		Field: "EndDate",
    // 		Op: "$lte",
    // 		Value: moment(data.search.DateTo).utc().format("YYYY-MM-DDT23:59:00Z"),
    // 	});
    // }

    // if (filters.length == 1) {
    // 	payload = {
    // 		Where: filters[0],
    // 	};
    // } else if (filters.length > 1) {
    // 	payload = {
    // 		Where: {Op: "$and", Items: filters},
    // 	};
    // }
    return (
      await axios.post(`/sdp/unitcalendar/site/gets`, {
        Where: {
          Op: "$and",
          Items: [
            {
              Op: "$eq",
              field: "Dimension.Key",
              value: "Site",
            },
            {
              Op: "$in",
              field: "Dimension.Value",
              value: site,
            },
          ],
        },
      })
    ).data;
  } catch (error) {
    util.showError(error);
  }
}

async function FindUnitCalendarsSORefNo(SORefNo) {
  try {
    return (
      await axios.post(`/sdp/unitcalendar/find`, {
        Where: {
          Op: "$eq",
          Field: "SORefNo",
          Value: SORefNo,
        },
      })
    ).data;
  } catch (error) {
    util.showError(error);
  }
}

async function GetSO(_id) {
  try {
    return (await axios.post(`/sdp/salesorder/get`, [_id])).data;
  } catch (error) {
    util.showError(error);
  }
  return undefined;
}

async function GetAsset(_ids) {
  try {
    return (
      await axios.post(`/tenant/asset/find`, {
        Select: ["_id", "Name"],
        Where: {
          OP: "$in",
          Field: "_id",
          Value: _ids,
        },
      })
    ).data;
  } catch (error) {}
}

async function GetSiteDimension(_ids) {
  try {
    return (
      await axios.post(`/tenant/dimension/find?DimensionType=Site`, {
        Where: {
          OP: "$in",
          Field: "_id",
          Value: _ids,
        },
      })
    ).data;
  } catch (error) {
    util.showError(error);
  }
  return undefined;
}

function refreshData(name, v1, v2) {
  util.nextTickN(2, () => {
    data.Projects = [];
    data.Skip = 0;
    Projects();
  });
}

onBeforeMount(() => {
  Init();
});

onMounted(() => {
  // window.addEventListener("scroll", onScrollVertical);
  // const app = document.querySelector("#app");
  // app.addEventListener("scroll", onScrollVertical);
});

// function OnClickDropDown(clickedproject) {
// 	data.Projects = data.Projects.map((project) => {
// 		if (clickedproject._id == project._id) {
// 			if (!project.visible) {
// 				GetProjectWithSites(project._id);
// 			} else {
// 				project.Lines = undefined;
// 			}
// 			return {
// 				...project,
// 				visible: !project.visible,
// 			};
// 		}
// 		return {
// 			...project,
// 		};
// 	});
// }

const getDataCustomer = async (_id) => {
  try {
    const dataresponse = await axios.post(`/tenant/customer/get`, [_id]);
    return dataresponse.data;
  } catch (error) {
    util.showError(error);
  }
};

async function FindSpecVariants(_ids) {
  try {
    const dataresponse = await axios.post(`/tenant/specvariant/find`, {
      Where: {
        Op: "$in",
        Field: "_id",
        Value: _ids,
      },
      Select: ["Name", "_id"],
    });

    return dataresponse.data;
  } catch (error) {
    util.showError(error);
  }
}

function onScrollVertical(event) {
  if (
    event.target.scrollTop + event.target.clientHeight >=
    event.target.scrollHeight
  ) {
    data.Skip += 1;
    Projects();
  }
}
</script>

<style>
.sticky-row {
  height: 10px;
  min-height: 10px;
  max-height: 10px;
  position: sticky;
  top: 0px;
  z-index: 2;
  background-color: white;
}

/* .tbl th {
	position: sticky;
	top: 0px;
	border-width: 1px;
	border-style: solid;
	border-color: black;
	z-index: 1;
	background-color: white;
} 
*/
.sticky-col {
  position: sticky;
  background-color: white;
  width: 150px;
  min-width: 150px;
  max-width: 150px;
  /* z-index: 1; */
  left: 0px;
}
</style>
