import "@coreui/coreui/dist/css/coreui.min.css"
import "./assets/main.css"

import { createPinia } from "pinia"
import { createApp } from "vue"

import App from "./App.vue"
import router from "./router"

// https://docs.fontawesome.com/web/use-with/vue/add-icons
/* import the fontawesome core */
import { library } from "@fortawesome/fontawesome-svg-core"
import { faHome } from "@fortawesome/free-solid-svg-icons"
/* import font awesome icon component */
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome"

/* import specific icons */
library.add(faHome)

const app = createApp(App)
app.component("font-awesome-icon", FontAwesomeIcon)
app.use(createPinia())
app.use(router)

Promise.resolve().then(() => {
  app.mount("#app")
})