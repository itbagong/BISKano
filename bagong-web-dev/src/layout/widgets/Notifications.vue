<template>
  <Menu as="div" class="relative inline-block text-left">
    <MenuButton class="hover:text-primary cursor-pointer relative">
      <mdicon name="bell"></mdicon>
      <span
        v-if="data.count > 0"
        class="absolute top-0 right-[-5px] flex items-center justify-center h-4 w-4 rounded-full bg-red-500 text-white text-xxs px-1"
      >
        {{ data.count }}
      </span>
    </MenuButton>
    <Transition
      enter-active-class="transition ease-out duration-100"
      enter-from-class="transform opacity-0 scale-95"
      enter-to-class="transform opacity-100 scale-100"
      leave-active-class="transition ease-in duration-75"
      leave-from-class="transform opacity-100 scale-100"
      leave-to-class="transform opacity-0 scale-95"
    >
      <MenuItems>
        <div
          ref="menuContainer"
          class="scroll-container origin-top-right absolute right-0 mt-2 w-96 max-h-96 overflow-y-auto rounded-md shadow-lg bg-white ring-1 ring-black ring-opacity-5 focus:outline-none"
          @scroll="handleScroll"
        >
          <div class="">
            <MenuItem v-if="data.notifications.length === 0">
              <div class="px-6 py-2 text-gray-700">No notifications</div>
            </MenuItem>
            <template v-else>
              <MenuItem
                v-slot="{ active }"
                v-for="notification in data.notifications"
                :key="notification._id"
              >
                <a href="#" @click="openLink(notification)">
                  <div
                    class="relative px-6 py-2 border-b border-gray-200 hover:bg-slate-100"
                  >
                    <mdicon
                      v-if="!notification.IsRead"
                      name="circle-medium"
                      class="text-primary absolute top-1/3 left-0"
                    ></mdicon>
                    <div class="flex items-center gap-2">
                      <div
                        v-if="notification.IsApproval"
                        class="rounded-md bg-orange-500 py-0.5 px-2.5 border border-transparent text-sm text-white transition-all shadow-sm"
                      >
                        Approval
                      </div>
                      <span v-if="!notification.IsApproval" class="text-sm text-gray-500 italic">{{
                        notification.Message
                      }}</span>
                    </div>
                    <h3
                      class="text-sm"
                      :class="notification.IsRead ? 'font-medium' : 'font-bold'"
                    >
                      {{ notification.Menu }} - {{ notification.TrxType }}
                    </h3>
                    <p class="text-xs mb-2 text-gray-500">
                      {{ moment(notification.TrxDate).format("DD MMMM YYYY") }}
                    </p>
                    <p class="text-xs">{{ notification.Text }}</p>
                  </div>
                </a>
              </MenuItem>
            </template>
            <MenuItem v-if="data.loading">
              <div class="px-6 py-2 text-gray-700">
                <loader kind="skeleton" skeleton-kind="list" />
              </div>
            </MenuItem>
          </div>
        </div>
      </MenuItems>
    </Transition>
  </Menu>
</template>
<script setup>
import { authStore } from "@/stores/auth";
import { Menu, MenuButton, MenuItem, MenuItems } from "@headlessui/vue";
import { util } from "suimjs";
import { inject, ref, reactive, onMounted, watch } from "vue";
import Loader from "@/components/common/Loader.vue";
import moment from "moment";
import { useRouter } from "vue-router";

const router = useRouter();
const auth = authStore();
const axios = inject("axios");
const menuContainer = ref(null);

const data = reactive({
  count: 0,
  notifications: [],
  loading: false,
  filter: {
    Skip: 0,
    Take: 10,
  },
  currentPage: 1,
  isAllLoaded: false,
  auth: auth,
});

function openLink(item) {
  axios.post("/fico/notification/save", { ID: item._id }).then((r) => {
    data.filter = {
      Skip: 0,
      Take: 10,
    };
    data.isAllLoaded = false;
    getNotifCount();
    getNotif(data.filter, true);
    util.nextTickN(2, () => {
      redirect(item.JournalID, item.TrxType, item.Menu);
    });
  });
}
function redirect(TrxID, TrxType, Menu) {
  let routeOpt = {};
  switch (TrxType) {
    case "Other":
      routeOpt = {
        name: "fico-LedgerJournal",
        query: { id: TrxID },
      };
      break;
    case "General Submission":
      routeOpt = {
        name: "bagong-SubmissionVendor",
        query: { id: TrxID },
      };
      break;
    case "Employee Expense":
      routeOpt = {
        name: "bagong-SubmissionEmployeeExpense",
        query: { id: TrxID },
      };
      break;
    case "Site Entry Expense":
    case "Vendor Purchase":
    case "Credit Note":
    case "Good Receive":
    case "Site Entry Expense":
    case "Employee Expense":
    case "General Submission":
      routeOpt = {
        name: "fico-VendorTransaction",
        query: { id: TrxID },
      };
      break;
    case "Site Entry Income":
    case "Customer Sales":
    case "Credit Note":
    case "Customer Deposit":
    case "Mining Invoice - Rent":
    case "Trayek Invoice":
    case "BTS Invoice":
    case "General Invoice":
    case "General Invoice - Tourism":
    case "General Invoice - Sparepart":
      routeOpt = {
        name: "fico-CustomerTransaction",
        query: { id: TrxID },
      };
      break;
    case "CASH IN":
      routeOpt = {
        name: "bagong-CashIn",
        query: { trxid: TrxID },
      };
      if (Menu === "Petty Cash Submission") {
        routeOpt = {
          name: "bagong-SubmissionPettyCash",
          query: { trxid: TrxID },
        };
      }
      break;
    case "SUBMISSION CASH IN":
      routeOpt = {
        name: "bagong-CashIn",
        query: { trxid: TrxID, id: "SubmissionCashIn" },
      };
      break;
    case "CASH OUT":
      routeOpt = {
        name: "bagong-CashOut",
        query: { trxid: TrxID },
      };
      break;
    case "SUBMISSION CASH OUT":
      routeOpt = {
        name: "bagong-CashOut",
        query: { trxid: TrxID, id: "SubmissionCashIn" },
      };
      break;
    case "Purchase Request":
      routeOpt = {
        name: "scm-PurchaseRequest",
        query: { trxid: TrxID },
      };
      break;
    case "Purchase Order":
      routeOpt = {
        name: "scm-PurchaseOrder",
        query: { trxid: TrxID },
      };
      break;
    case "Movement In":
      routeOpt = {
        name: "scm-InventoryJournal",
        query: { trxid: TrxID, type: "Movement In", title: "Movement In" },
      };
      break;
    case "Movement Out":
      routeOpt = {
        name: "scm-InventoryJournal",
        query: { trxid: TrxID, type: "Movement Out", title: "Movement Out" },
      };
      break;
    case "Transfer":
      routeOpt = {
        name: "scm-InventoryJournal",
        query: { trxid: TrxID, type: "Item Transfer", title: "Item Transfer" },
      };
      break;
    case "Item Request":
      routeOpt = {
        name: "scm-ItemRequest",
        query: { id: TrxID },
      };
      break;
    case "Inventory Receive":
      routeOpt = {
        name: "scm-InventTrx",
        query: {
          trxid: TrxID,
          type: "Inventory Receive",
          title: "Inventory Receive",
        },
      };
      break;
    case "Inventory Issuance":
      routeOpt = {
        name: "scm-InventTrx",
        query: {
          trxid: TrxID,
          type: "Inventory Issuance",
          title: "Inventory Issuance",
        },
      };
      break;
    case "Work Request":
      routeOpt = {
        name: "mfg-WorkRequestor",
        query: { id: TrxID },
      };
      break;
    case "Work Order":
      routeOpt = {
        name: "mfg-WorkOrderPlan",
        query: { id: TrxID },
      };
      break;
    case "Item":
    case "Asset":
      if (Menu === "Sales Quotation") {
        routeOpt = {
          name: "sdp-salesquotation",
          query: { id: TrxID },
        };
      } else if (Menu === "Sales Order") {
        routeOpt = {
          name: "sdp-SalesOrder",
          query: { id: TrxID },
        };
      }
      break;
    default:
      break;
  }
  if (routeOpt?.name) {
    const url = router.resolve(routeOpt);
    window.open(url.href, "_blank");
  }
}
function getNotifCount() {
  axios.post("/fico/notification/count").then((r) => {
    data.count = r.data.count;
  });
}
function getNotif(filter, isreset) {
  data.loading = true;
  const payload = filter;
  axios
    .post("/fico/notification/shows", payload)
    .then(
      (r) => {
        if (r.data.data.length === 0) {
          data.isAllLoaded = true;
        }
        if (isreset) {
          data.notifications = r.data.data;
        } else {
          data.notifications = [...data.notifications, ...r.data.data];
        }
      },
      (e) => util.showError(e)
    )
    .finally(() => {
      util.nextTickN(2, () => {
        data.loading = false;
      });
    });
}
const handleScroll = () => {
  const element = menuContainer.value;
  if (element.scrollTop + element.clientHeight >= element.scrollHeight - 10) {
    if (!data.loading && !data.isAllLoaded) {
      data.currentPage += 1;
      data.filter.Skip = (data.currentPage - 1) * data.filter.Take;
      getNotif(data.filter);
    }
  }
};
function init() {
  getNotifCount();
  data.filter = {
    Skip: 0,
    Take: 10,
  };
  data.isAllLoaded = false;
  getNotif(data.filter, true);
}
watch(
  () => auth.appData?.CompanyID,
  (nv) => {
    init();
  },
  { deep: true }
);
onMounted(() => {
  init();
});
</script>
<style scoped>
.text-xxs {
  font-size: 0.5rem;
  line-height: 0.75rem;
}
.scroll-container::-webkit-scrollbar {
  display: none;
}

.scroll-container {
  -ms-overflow-style: none; /* IE and Edge */
  scrollbar-width: none; /* Firefox */
}
</style>
