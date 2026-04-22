import {h} from "vue";
import DefaultTheme from "vitepress/theme";
import HeroInstallTabs from "./components/HeroInstallTabs.vue";
import "./custom.css";

export default {
  extends: DefaultTheme,
  Layout: () => {
    return h(DefaultTheme.Layout, null, {
      "home-hero-info-after": () => h(HeroInstallTabs),
    });
  },
  enhanceApp({app}) {
    app.component("HeroInstallTabs", HeroInstallTabs);
  },
};
