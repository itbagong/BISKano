<template>
  <!-- <div
    class="bg-cover w-full h-screen [&>*]:py-4 [&>*]:px-8 clear_layout"
    :style="{ backgroundImage: `url(${bgLogin})` }"
  >
    <div class="flex justify-start h-[70px]">
      <img src="@/assets/img/bagong.png" width="150" />
    </div>
    <div class="flex justify-end mt-[-70px]">
      <s-button
        label="Wistle Blower"
        class="wb_btn"
        @click="data.showBW = true"
      />
    </div> -->
    <RouterView />
    <!-- <s-modal :display="data.showBW" hide-title hide-buttons>
      <div class="card">
        <div class="flex justify-end">
          <mdicon
            width="22"
            name="close"
            class="modal_action cursor-pointer"
            @click="data.showBW = false"
          />
        </div>
        <data-list
          ref="listControl"
          hide-title
          no-gap
          init-app-mode="form"
          grid-mode="form"
          new-record-type="form"
          :form-config="'/she/wistleblower/formconfig'"
          form-insert="/she/wistleblower/save"
          form-update="/she/wistleblower/save"
          form-hide-cancel
          form-hide-submit
        >
          <template #form_footer_1="{ item }">
            <div class="w-full flex justify-end">
              <s-button label="Submit" class="wb_btn" @click="onSubmit(item)" />
            </div>
          </template>
        </data-list>
      </div>
    </s-modal>
  </div> -->
</template>
<script setup>
// import bgLogin from "@/assets/img/bg-login.png";
// import { SButton, SModal, DataList } from "suimjs";
// import { reactive, inject, ref } from "vue";

// const axios = inject("axios");
// const listControl = ref(null);

// const data = reactive({
//   showBW: false,
// });

// function onSubmit(item) {
//   if (item["Date"] == undefined) {
//     item["Date"] = new Date();
//   }
//   listControl.value.submitForm(
//     item,
//     () => {},
//     () => {}
//   );
//   data.showBW = false;
// }
</script> 
