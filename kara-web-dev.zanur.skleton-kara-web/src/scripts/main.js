import { createApp } from 'vue'
import App from '@/App.vue'
import { createPinia } from 'pinia'
import router from '@/router/main-router'
import '@/index.css'
//import util from '@/assets/js/util.js'

// mdi
import mdiVue from 'mdi-vue/v3'
import * as mdijs from '@mdi/js'

//-- axios
import axios from 'axios'
import VueAxios from 'vue-axios'

//-- vue-select
import 'vue-select/dist/vue-select.css';
import vSelect from 'vue-select'

//-- vue-cookies
import VueCookies from 'vue3-cookies'

//-- vue quill
import { QuillEditor } from '@vueup/vue-quill';
import '@vueup/vue-quill/dist/vue-quill.snow.css';

//-- notif
import { notifStore } from '@/stores/notif.js'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(VueCookies, {
    expireTimes: "7d",
    path: import.meta.env.VITE_COOKIE_PATH,
    domain: import.meta.env.VITE_COOKIE_HOST,
    secure: import.meta.env.VITE_COOKIE_SECURE || false
})

//-- load mdi icon
app.use(mdiVue, { icons: mdijs })

//-- auth
import { authStore } from '@/stores/auth.js'
import { util } from 'suimjs'
import VueSelect from 'vue-select'

//-- prepare axios setting and loading it into vue
const api_url = import.meta.env.VITE_API_URL
const apiAxios = axios.create({
    baseURL: api_url
})

apiAxios.interceptors.request.use(function (cfg) {
    const auth = authStore()
    if (auth.appToken != '') cfg.headers.common['Authorization'] = 'Bearer ' + auth.appToken
    return cfg
})

apiAxios.interceptors.response.use(r => {
    if (r.data && r.data.error) {
        const errorMessage = r.data.error
        if (errorMessage == "EOF") {
            return Promise.reject("Sorry, but the data you looking for is not exist")
        } else if (errorMessage.indexOf("invalid access token") > 0 || errorMessage.indexOf("invalid token") > 0) {
            //auth.clear()
        }
        return Promise.reject(errorMessage)
    }
    return Promise.resolve(r)
}, e => {
    return Promise.reject(e)
})

app.use(VueAxios, axios)
app.provide('axios', apiAxios)

app.mount('#app')

//-- global components
app.component('v-select', vSelect)
app.component('QuillEditor', QuillEditor)

//-- notif
const notif = notifStore()
util.setNotifStore(notif)