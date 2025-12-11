import { defineStore } from "pinia";
export const layoutStore = defineStore("layout", {
  state: () => ({
    name: "clear",
    firstPath: "/",
    breadcrumbs: [],
    breakPoint: "",
    miniNav: true,
    addDataMaster: false,
    appID: "bagong",
    headOfficeID: "SITE020",
  }),

  actions: {
    change(name) {
      this.name = name;
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
      this.miniNav = miniNav;
    },
    setAddDataMaster(addDataMaster) {
      this.addDataMaster = addDataMaster;
    },
  },

  getters: {
    path(state) {
      return "/fico";
    },
    homePage(state) {
      return "/";
    },
    loginPage(state) {
      return "/iam/Login";
    },
  },
});
