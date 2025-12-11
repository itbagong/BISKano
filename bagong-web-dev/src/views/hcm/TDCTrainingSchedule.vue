<template>
  <div class="w-full">
    <div class="flex flex-col suim_card show_title w-full bg-white suim_datalist card">
      <div class="flex">
        <!-- Bagian Kiri: Kalender -->
        <div class="w-3/4 p-4">
          <FullCalendar
            ref="calendarRef"
            :options="calendarOptions"
            @eventClick="handleEventClick"
          />
        </div>

        <div
          v-if="isLoading"
          class="absolute inset-0 flex items-center justify-center bg-gray-500 bg-opacity-50 z-10"
        >
          <div class="loader"></div> <!-- Ganti dengan spinner -->
        </div>
    
        <!-- Bagian Kanan: Detail Event -->
        <div class="w-1/4 p-4 border-l">
          <h2 class="text-lg font-bold mb-2">Schedule</h2>
          <div v-show="data.selectedEvent.title">
            <span class="event-title">{{ data.selectedEvent.title }}</span>
            <br />
            <span class="event-location">{{ data.selectedEvent.AssessmentType }}</span>
            <br />
            <span class="event-time">Date From: {{ moment(data.selectedEvent.TrainingDateFrom).format("YYYY-MM-DD").toString() }}</span>
            <br />
            <span class="event-time">Date To: {{ moment(data.selectedEvent.TrainingDateTo).format("YYYY-MM-DD").toString() }}</span>
          </div>
          <div v-show="data.selectedEvent.title == undefined">
            <p>No event selected</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { layoutStore } from "@/stores/layout.js";
import { ref, inject, onMounted, onUnmounted, reactive } from "vue";
import FullCalendar from "@fullcalendar/vue3";
import dayGridPlugin from "@fullcalendar/daygrid";
import timeGridPlugin from "@fullcalendar/timegrid";
import interactionPlugin from "@fullcalendar/interaction";
import moment from "moment";
import { util } from "suimjs";

const axios = inject("axios");

layoutStore().name = "tenant";

// const selectedEvent = ref(null);
const calendarRef = ref(null);
const isLoading = ref(false);

const data = reactive({
  mode: 'MONTH', //MONTH||YEAR
  selectedEvent: {},
});

function handleEventClick(info) {
  data.selectedEvent = {
    title: info.event.title,
    AssessmentType: info.event.extendedProps.AssessmentType,
    TrainingDateFrom: info.event.extendedProps.TrainingDateFrom,
    TrainingDateTo: info.event.extendedProps.TrainingDateTo,
    description: info.event.extendedProps.description,
  }
}

function handleDateClick(params) {
  // console.log(params)
}

async function fetchSchedules(date) {
  isLoading.value = true; 
  const param = {
    Date: date,
    ShowBy: data.mode
  };
  try {
    const response = await axios.post(`/hcm/tdc/get-schedules`, param);
    const schedules = response.data.map((item, idx) => ({
      title: item.Title,
      start: moment(item.Start).format("YYYY-MM-DDTHH:mm:ss"),
      end: moment(item.End).add(1, "hours").format("YYYY-MM-DDTHH:mm:ss"),
      description: item.Description,
      AssessmentType: item.AssessmentType,
      TrainingDateFrom: item.TrainingDateFrom,
      TrainingDateTo: item.TrainingDateTo,
      className:
        item.AssessmentType == "Assessment Staff"
          ? "event-tc-staff"
          : item.AssessmentType == "Assessment Driver"
          ? "event-tc-driver"
          : "event-tc-mechanic",
    }));
    calendarOptions.events = schedules;
  } catch (error) {
    util.showError(error);
  } finally {
    isLoading.value = false;
  }
}

function handleDatesSet(info) {
  const input = info.view.title;
  const [month, year] = input.split(" ");
  let date = new Date();
  if(year && month){
    data.mode = "MONTH";
    const monthIndex = new Date(`${month} 1`).getMonth();
    console.log("monthIndex: ", monthIndex)
    date = new Date(Date.UTC(year, monthIndex, 1, 0, 0, 0));
  } else {
    data.mode = "YEAR";
    let monthIndex = new Date().getMonth();
    console.log("monthIndex: ", monthIndex)
    date = new Date(Date.UTC(month, monthIndex, 1, 0, 0, 0));
  }
  console.log("date: ", date)
  const isoDate = date.toISOString();
  fetchSchedules(isoDate);
}

const calendarOptions = reactive({
  plugins: [dayGridPlugin, timeGridPlugin, interactionPlugin],
  initialView: "dayGridMonth",
  headerToolbar: {
    left: "today",
    center: "prev,title,next",
    right: "dayGridMonth,dayGridYear",
  },
  events: [],
  datesSet: handleDatesSet,
  eventClick: handleEventClick,
  eventContent: function(arg) {
    return {
      html: `<div class="fc-event-title">${arg.event.title}</div>`, 
    };
  },
});

onMounted(() => {
  if (calendarRef.value) {
    // console.log("Calendar instance: ", calendarRef.value.getApi());
  }
});

onUnmounted(() => {
  if (calendarRef.value) {
    // console.log("Destroying FullCalendar...");
    calendarRef.value.getApi().destroy();
  }
});
</script>

<style>
div.fc-header-toolbar.fc-toolbar.fc-toolbar-ltr > div:nth-child(2) > div {
  display: flex;
  gap: 20px;
}
.fc .fc-button-primary.fc-button-active{
  color: #40444B !important;
}
.fc .fc-prev-button.fc-button-primary, .fc .fc-next-button.fc-button-primary {
  border-radius: 10px !important;
}
.fc .fc-button-primary {
  background-color: white !important;
  color: #2c3e50 !important;
  border: 1px solid #2c3e50 !important;
  text-transform: capitalize;
}

.fc-event.event-tc-staff {
  background-color: #F0EBFC !important; 
  color: #8E6BD5 !important; 
  border: none !important;
}

.fc-event.event-tc-driver {
  background-color: #DBF4E9 !important;
  color: #10A55F !important; 
  border: none !important;
}

.fc-event.event-tc-mechanic {
  background-color: #FEEEDA !important; 
  color: #DE8208 !important; 
  border: none !important;
} 

.fc-daygrid-event-dot {
  display: none !important;
}

.fc-event-title {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-align: center;
  padding: 1px 10px;
}

.fc .fc-view-harness {
  z-index: 0;
}
</style>

<style scoped>
.event-title {
  color: #40444B;
  font-size: 16px;
}
.event-time {
  color: #8A93A3;
  font-size: 14px;
}
.event-location {
  /* color: #8A93A3; */
  font-size: 14px;
}
.loader {
  border: 4px solid rgba(255, 255, 255, 0.2);
  border-left-color: #ffffff;
  border-radius: 50%;
  width: 40px;
  height: 40px;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.bg-gray-500.bg-opacity-50 {
  background-color: rgba(75, 85, 99, 0.5); /* Abu-abu dengan transparansi */
}
</style>