<template>
<div
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
    </div>
    <div class="flex items-center h-[calc(100vh-70px)] w-full page_login">
      <div class="flex w-full">
        <div class="md:basis-1/3 basis-1/1">
          <div class="text-white text-3xl font-bold mb-4">
            Your Safety and Satisfaction is Our Priority
          </div>
          <div class="text-xl text-white mb-4">
            We make your transportation easy and simple!
          </div>
          <s-button
            label="Login"
            class="login_btn"
            @click="data.showLogin = true"
          />
           <div class="flex flex-col gap-2 mt-4  pb-5">
                <a
                  class="text-white cursor-pointer"
                  @click="router.push('/iam/publicresetpassword')"
                  >I forget my credentials, please help me</a
                >
                <a
                  class="text-white cursor-pointer" @click="router.push('/iam/register')"
                  >I don't have account and want to create new one</a
                >
              </div>
        </div>
      </div>
      <s-modal :display="data.showLogin" hide-title hide-buttons>
        <div class="flex justify-end">
          <mdicon
            width="22"
            name="close"
            class="modal_action cursor-pointer"
            @click="data.showLogin = false"
          />
        </div>
        <div class="card relative">
          <s-form
            class="sFormLogin"
            ref="loginFormCtl"
            v-model="data.record"
            :config="data.formCfg"
            keep-label
            only-icon-top
            buttons-on-bottom
            :buttons-on-top="false"
            hide-cancel
            submit-text="Login"
            submit-icon="login-variant"
            @submitForm="login"
            auto-focus
          >
            <template #section_General_header>
              <div class="w-full flex justify-center mb-4">
                <img src="@/assets/img/logofull-lg.png" width="180" />
              </div>
            </template>
            <template #footer_2>
              <div class="flex flex-col gap-2 mt-4 items-center pb-5">
                <a
                  class="btn_text_primary"
                  @click="router.push('/iam/publicresetpassword')"
                  >I forget my credentials, please help me</a
                >
                <a class="btn_text_primary" @click="router.push('/iam/register')"
                  >I don't have account and want to create new one</a
                >
              </div>
            </template>
          </s-form>
        </div>
      </s-modal>
    </div>
    <s-modal :display="data.showBW" hide-title hide-buttons>
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
              <s-button label="Submit" class="wb_btn" @click="onSubmitBW(item)" />
            </div>
          </template>
        </data-list>
      </div>
    </s-modal>
</div>
</template>

<script setup>
import bgLogin from "@/assets/img/bg-login.png";
import { layoutStore } from "@/stores/layout";
import { onMounted, reactive, inject, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { authStore } from "@/stores/auth";
import {
  formInput,
  createFormConfig,
  SCard,
  SForm,
  SButton,
  SModal,
  DataList
} from "suimjs";
import { notifStore } from "@/stores/notif";

layoutStore().change("clear");

const loginFormCtl = ref(null);
const notif = notifStore();

const data = reactive({
  formCfg: {},
  record: {},
  showLogin: false,
  showBW: false,
});

const axios = inject("axios");
const router = useRouter();
const auth = authStore();

function login(record, cb1, cb2) {
  axios
    .post(
      "iam/http-auth",
      { CheckName: "LoginID" }, //-- timeout: 6 hours
      { auth: { username: record.LoginID, password: record.Password } }
    )
    .then(
      (r) => {
        if (r.data.Data.Use2FA && r.data.Data.Status != "Expired") {
          router.push("/validate2fa");
          cb1();
        } else { 
          auth.updateJwt(r.data);
          router.push("/");
          cb1();
        }
      },
      (e) => {
        cb2();
        notif.add({ kind: "error", message: e });
      }
    );
}


onMounted(() => {
  const cfg = createFormConfig("Login Form", true);
  cfg.setting.showTitle = false;
  const loginID_input = new formInput();
  loginID_input.field = "LoginID";
  loginID_input.label = "Login ID";
  loginID_input.kind = "string";
  loginID_input.required = true;

  const loginPassword_input = new formInput();
  loginPassword_input.field = "Password";
  loginPassword_input.label = "Password";
  loginPassword_input.kind = "password";
  loginPassword_input.required = true;

  cfg
    .addSection("General", false)
    .addRowAuto(1, loginID_input, loginPassword_input);
  data.formCfg = cfg.generateConfig();
});
function onSubmitBW(item) {
  if (item["Date"] == undefined) {
    item["Date"] = new Date();
  }
  listControl.value.submitForm(
    item,
    () => {},
    () => {}
  );
  data.showBW = false;
}
</script>
<style>
.wb_btn {
  @apply bg-red-700 disabled:bg-slate-500 disabled:text-white disabled:cursor-not-allowed  text-[12px] font-bold text-white;
  padding: 5px 10px !important;
}
.page_login .login_btn {
  @apply bg-amber-400   disabled:bg-slate-500 disabled:text-white disabled:cursor-not-allowed  text-[1rem] font-bold text-white;
  /* padding: 0px 40px !important; */
  width: max-content;
}
.page_login .login_btn > button{
  padding: 0px 40px !important;

}

.sFormLogin .submit_btn {
  @apply text-amber-300;
  justify-content: center;
  width: 100% !important;
  background: #40444b;
}
.sFormLogin .submit_btn button{
  @apply justify-center;
  width: 100% !important; 
}

.sFormLogin .suim_form_button.form_button_bottom > div > div > div {
  display: none !important;
}
.sFormLogin input {
  @apply py-[6px] border border-slate-500 rounded-[4px];
}
/* .sFormLogin .section_group_container +  div .grow{
    display: none !important;
} 
.sFormLogin .section_group_container +  div button   {
    justify-content: center;
    width:  100% !important;
} */
</style>

<style scoped>
.card {
  background: white;
  border-radius: 0;
  box-shadow: none;
  border: none;
}
</style>
