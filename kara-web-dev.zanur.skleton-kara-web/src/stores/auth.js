import { defineStore } from "pinia"
import { inject } from "vue"
import { useCookies } from 'vue3-cookies'
import router from "@/router/main-router"
import moment from "moment"
import { useRoute } from "vue-router"

function vcToState(vc) {
    const auth = vc.get('auth');
    if (auth) {
        return auth;
    }

    return {
        appToken: '',
        appData: undefined,
        tokenExpiry: undefined,
    };
}

function stateToVc(state, vc) {
    vc.set('auth', {
        appToken: state.appToken,
        appData: state.appData,
        tokenExpiry: state.tokenExpiry
    })
}

export const authStore = defineStore('auth', {
    state: () => {
        const vc = useCookies().cookies
        return vcToState(vc)
    },

    actions: {
        set(userAuth) {
            const vc = useCookies().cookies
            const state = this
            state.appToken = userAuth.appToken
            state.appData = userAuth.appData
            state.tokenExpiry = userAuth.tokenExpiry
            stateToVc(state, vc)
        },

        updateJwt(jwt) {
            const vc = useCookies().cookies
            const state = this
            state.appToken = jwt.Token
            state.appData = jwt.Data
            state.tokenExpiry = jwt.ExpireTime
            stateToVc(state, vc)
        },

        setActivate(activated) {
            const vc = useCookies().cookies
            const state = this
            state.activated = activated
            stateToVc(state, vc)
        },

        mustNotLogin () {
            if (this.appToken && this.appToken!='') {
                router.push({path:'/general/noaccess'});
            }
        },
 
        hasAccess(checkModel) {
            const state = this;
            
            //-- TODO: check expiry, if yes, clear the jwt and redirect to nosession page
            if (moment(this.tokenExpiry).isBefore(new Date())) {
                router.push({path:'/login'});
            }
            
            const axios = inject('axios');
            const route = useRoute();
            axios.post('/iam/access/get', checkModel).then(
                r => { /*-- do nothing --*/ },
                e => {
                    router.push({path:'/general/noaccess?route=' + route.fullPath});
                });
        },

        clear() {
            const vc = useCookies().cookies
            const state = this
            state.appToken = ''
            state.appData = undefined
            stateToVc(state, vc)
        }
    }
})
