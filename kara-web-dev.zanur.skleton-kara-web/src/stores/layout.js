import { defineStore } from 'pinia';
export const layoutStore = defineStore("layout", {
    state: () => ({
        name: "clear",
        firstPath: "/",
        breadcrumbs: [],
        breakPoint: "",
        miniNav: true
    }),

    actions: {
        change(name) {
            this.name = name
        },
        setFirstPath(firstPath) {
            this.firstPath = firstPath;
        },
        setBreadCrumbs(breadcrumbs) {
            this.breadcrumbs = breadcrumbs;
        },
        setBreakPoint(breakPoint) {
            this.breakPoint = breakPoint;
        },
        setMiniNav(miniNav) {
            this.miniNav = miniNav
        }
    },

    getters: {
        path(state) {
            return "/app";
        },
        homePage(state) {
            return "/"
        },
        loginPage(state) {
            return "/login";
        },
    },
})
