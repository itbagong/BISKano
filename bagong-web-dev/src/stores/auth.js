import { defineStore } from "pinia"
import { inject } from "vue" 
import router from "@/router/main-router"
import moment from "moment"
import { useRoute } from "vue-router"
import { layoutStore } from "@/stores/layout";

function lsToState(){
    const auth =  JSON.parse(localStorage.getItem("xibarAuth"))
    if(auth){
        return auth
    }
    return {
        appToken: '',
        appData: undefined,
        tokenExpiry: undefined,
        companyId:'',
    };
}

function stateToLs(state) {
    localStorage.setItem('xibarAuth', JSON.stringify({
        appToken: state.appToken,
        appData: state.appData,
        tokenExpiry: state.tokenExpiry,  
        companyId:state.companyId
    }))
}
export const authStore = defineStore('auth', {
    state: () => { 
        return lsToState()
    },

    actions: {
        set(userAuth) { 
            const state = this
            state.appToken = userAuth.appToken
            state.appData = userAuth.appData
            state.tokenExpiry = userAuth.tokenExpiry
            stateToLs(state)
        },

        updateJwt(jwt) { 
            const state = this
            state.appToken = jwt.Token
            state.appData = jwt.Data
            state.tokenExpiry = jwt.ExpireTime
            stateToLs(state)
        },

        setActivate(activated) { 
            const state = this
            state.activated = activated
            stateToLs(state)
        },
        setCompanyId(companyId){ 
            const state = this
            state.appData.CompanyID = companyId
            stateToLs(state)
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
                router.push({path:'/iam/Login'});
            }
            
            const axios = inject('axios');
            const route = useRoute();
            axios.post('/iam/access/get', checkModel).then(
                r => { /*-- do nothing --*/ },
                e => {
                    router.push({path:'/share/noaccess?route=' + route.fullPath});
                });
        },
        

        clear() { 
            const state = this
            state.appToken = ''
            state.appData = undefined
            state.companyId = ''
            state.tokenExpiry = undefined

            stateToLs(state)
        },

        hasFeatureID(featureID){ 
          const idx =  this.appData?.RBAC?.Access?.findIndex(e=> e.FeatureID === featureID)
          if(idx === -1) return false 
          return this.appData.RBAC.Access[idx].Level > 0
           

        },

        validate() {
            const state = this
            if (state.appToken != "") {
                if (state.tokenExpiry) {
                    if (moment(state.tokenExpiry).isBefore(new Date())) {
                        this.clear();
                        router.push({path: layoutStore().loginPage});
                    }
                }   
            }
        },
        
        getRBAC(featureID){
            const roleIDs = this.appData?.RBAC?.RoleIDs;
            if (roleIDs.indexOf('Administrators') >= 0) {
                return {
                    canRead: true,     // 1
                    canCreate: true,   // 2
                    canUpdate: true,   // 4
                    canDelete: true,   // 8
                    canPosting: true,  // 16
                    canSpecial1: true, // 32
                    canSpecial2: true, // 64
                    canSpecial3: true,  // 128
                    Dimension: []
                }
            }


            const r = {
                canRead: false,     // 1
                canCreate: false,   // 2
                canUpdate: false,   // 4
                canDelete: false,   // 8
                canPosting: false,  // 16
                canSpecial1: false, // 32
                canSpecial2: false, // 64
                canSpecial3: false  // 128
            }
       
            const access =  this.appData?.RBAC?.Access ?? []
            const idx = access.findIndex(e =>  e.FeatureID === featureID)
            if(idx === -1) return {...r, Dimension:[]}

            const myAcces = access[idx] 
    
            let level = myAcces.Level 
            
            let i = 0   
            while (level > 0) {
              const bit = level % 2;
              r[Object.keys(r)[i]] = bit > 0
              level = Math.floor(level / 2);
              i++
            }

            return {...r, Dimension:myAcces.Dimension} 
        }
    }
})
