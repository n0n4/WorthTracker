import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import Axios from 'axios'

var globaladdr = "localhost:3000";//"98.217.67.202:3000";//"localhost:3000";
Vue.prototype.$addr = globaladdr;
store.$addr = globaladdr;
Vue.prototype.$http = Axios;

Vue.config.productionTip = false

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
