<template>
    <div>
        <div id="metricsModal" class="modal modal-fixed-footer">
            <div v-if="deviceMetrics !== undefined" class="modal-content">
                <h4 v-if="deviceMetrics.name !== ''">{{ deviceMetrics.name }}</h4>
                <h4 v-if="deviceMetrics.name === ''">{{ capCase(deviceMetrics.color) }}</h4>
                <div class="row">
                    <div class="col s12">
                        <ul class="tabs">
                            <li class="tab col s3"><a class="active" href="#gravity">Gravity</a></li>
                            <li class="tab col s3"><a href="#temperature">Temperature</a></li>
                            <li class="tab col s3"><a href="#signal">Signal</a></li>
                            <li class="tab col s3"><a href="#battery">Battery</a></li>
                        </ul>
                    </div>
                    <div id="gravity" class="col s12">
                        <chart :data="dataForDeviceGravity(deviceMetrics)" :options="{responsive: false, maintainAspectRatio: false}" :width="400" :height="200"></chart>
                    </div>
                    <div id="temperature" class="col s12">Test 2</div>
                    <div id="signal" class="col s12">Test 3</div>
                    <div id="battery" class="col s12">Test 4</div>
                </div>
            </div>
            <div class="modal-footer">
                <a href="#!" class="modal-action modal-close waves-effect waves-green btn red">Close</a>
            </div>
        </div>
        <div id="settingsModal" class="modal modal-fixed-footer">
            <div class="modal-content">
                <h4 v-if="deviceSettings.name !== ''">{{ deviceSettings.name }} ({{ deviceSettings.color }})</h4>
                <h4 v-if="deviceSettings.name === ''">{{ capCase(deviceSettings.color) }}</h4>
                <div class="row">
                    <form class="col s12">
                        <div class="row">
                            <div class="input-field col s12">
                                <input v-model="deviceSettings.name" placeholder="FV5" id="name"
                                       type="text" class="validate">
                                <label for="name">Name/Location</label>
                            </div>
                        </div>
                        <div class="row">
                            <div class="input-field col s12">
                                <input v-model="deviceSettings.endpoint" id="endpoint"
                                       placeholder="https://brewmaker.com/metrics/M39dk2d93" type="text"
                                       class="validate">
                                <label for="endpoint">HTTP Endpoint (for posting stats)</label>
                            </div>
                        </div>
                    </form>
                </div>
            </div>
            <div class="modal-footer">
                <a @click="updateDevice(deviceSettings)" href="#!" class="modal-action waves-effect waves-green-lighten btn green">Save</a>
                <a href="#!" class="modal-action modal-close waves-effect waves-green btn red">Close</a>
            </div>
        </div>
        <!--nav class="light-blue">
            <div class="nav-wrapper">
                <a href="#" class="brand-logo center"><i class="mdi mdi-water"></i> Hydromonitor</a>
            </div>
        </nav-->
        <div v-for="device,i in devices" class="widget" :class="device.color">
            <div class="info name-wrapper">
                <div v-if="device.name !== ''" class="name">{{device.name}}</div>
                <div v-if="device.name === ''" class="name">{{capCase(device.color)}}</div>
                <div class="timestamp"><i class="mdi mdi-clock"></i> Updated {{ timeAgo(device.updated) }}</div>
            </div>
            <div class="info gravity-wrapper">
                <div class="gravity"><i class="mdi mdi-trending-down"></i> {{ device.latest_metrics.gravity }}</div>
                <!--div class="gravity-stats">{{ device.latest_metrics.gravityChangeInt }} ({{ device.latest_metrics.gravityChangePerc }})</div-->
                <div class="gravity-stats">-0.002 (-3.04%)</div>
            </div>
            <div class="info temperature">
                <span class='update'><i class="mdi mdi-thermometer"></i> {{ device.latest_metrics.temperature }}</span>
            </div>
            <div class="info battery">
                <span class='update'><i class="mdi" :class="getBatteryIcon(device.latest_metrics.battery)"></i> {{ device.latest_metrics.battery }}</span>
            </div>
            <div class="info signal">
                <span class='update'><i class="mdi mdi-signal"></i> {{ device.latest_metrics.power }}</span>
            </div>
            <div class="settings">
                <a class="waves-effect waves-light dropdown-button" href="#" :data-activates="'settingsDropdown'+i">
                    <i class="mdi mdi-dots-vertical"></i>
                </a>
                <ul :id="'settingsDropdown' + i" class='dropdown-content'>
                    <li><a @click="showGraphs(device)" href="#">Graphs</a></li>
                    <li><a @click="refreshDevice(device.id)" href="#">Refresh</a></li>
                    <li class="divider"></li>
                    <li><a @click="showSettings(device)" href="#">Settings</a></li>
                    <li><a @click="resetDevice(device.id)" href="#">Reset</a></li>
                    <li><a @click="deleteDevice(device.id)" href="#">Delete</a></li>
                </ul>
            </div>
        </div>
    </div>
</template>

<script>
  import Chart from './components/chart.vue'
  import moment from 'moment'
  export default {
    name: 'app',
    components: { Chart },
    data () {
      return {
        deviceMetrics: {},
        deviceSettings: {},
        devices: [
          {
            "id": "5c2320e0c41a4e238e4a22ca498fa439",
            "name": "FV5",
            "color": "blue",
            "endpoint": "",
            "disabled": false,
            "error": "context deadline exceeded",
            "latest_metrics": {
              "power": -65,
              "battery": 50,
              "temperature": 77,
              "gravity": 1.197,
              "created": "2017-07-13T00:25:37.164272351Z"
            },
            "created": "2017-07-12T03:31:26.940667846Z",
            "updated": "2017-07-12T22:50:56.642318342Z"
          },
          {
            "id": "5c2320e0c41a4e238e4a22ca498fa439",
            "name": "",
            "color": "black",
            "endpoint": "",
            "disabled": false,
            "error": "context deadline exceeded",
            "latest_metrics": {
              "power": -65,
              "battery": 50,
              "temperature": 77,
              "gravity": 1.197,
              "created": "2017-07-13T00:25:37.164272351Z"
            },
            "created": "2017-07-12T03:31:26.940667846Z",
            "updated": "2017-07-12T22:50:56.642318342Z"
          },
          {
            "id": "5c2320e0c41a4e238e4a22ca498fa439",
            "name": "",
            "color": "green",
            "endpoint": "",
            "disabled": false,
            "error": "context deadline exceeded",
            "latest_metrics": {
              "power": -65,
              "battery": 50,
              "temperature": 77,
              "gravity": 1.197,
              "created": "2017-07-13T00:25:37.164272351Z"
            },
            "created": "2017-07-12T03:31:26.940667846Z",
            "updated": "2017-07-12T22:50:56.642318342Z"
          },
          {
            "id": "5c2320e0c41a4e238e4a22ca498fa439",
            "name": "",
            "color": "orange",
            "endpoint": "",
            "disabled": false,
            "error": "context deadline exceeded",
            "latest_metrics": {
              "power": -65,
              "battery": 50,
              "temperature": 77,
              "gravity": 1.197,
              "created": "2017-07-13T00:25:37.164272351Z"
            },
            "created": "2017-07-12T03:31:26.940667846Z",
            "updated": "2017-07-12T22:50:56.642318342Z"
          },
          {
            "id": "5c2320e0c41a4e238e4a22ca498fa439",
            "name": "",
            "color": "pink",
            "endpoint": "",
            "disabled": false,
            "error": "context deadline exceeded",
            "latest_metrics": {
              "power": -65,
              "battery": 50,
              "temperature": 77,
              "gravity": 1.197,
              "created": "2017-07-13T00:25:37.164272351Z"
            },
            "created": "2017-07-12T03:31:26.940667846Z",
            "updated": "2017-07-12T22:50:56.642318342Z"
          },
          {
            "id": "5c2320e0c41a4e238e4a22ca498fa439",
            "name": "",
            "color": "purple",
            "endpoint": "",
            "disabled": false,
            "error": "context deadline exceeded",
            "latest_metrics": {
              "power": -65,
              "battery": 50,
              "temperature": 77,
              "gravity": 1.197,
              "created": "2017-07-13T00:25:37.164272351Z"
            },
            "created": "2017-07-12T03:31:26.940667846Z",
            "updated": "2017-07-12T22:50:56.642318342Z"
          },
        ],
        metrics: [
          {
            "power": -55,
            "battery": 50,
            "temperature": 77,
            "gravity": 1.024,
            "created": "2017-07-13T10:25:37.164272351Z"
          },
          {
            "power": -65,
            "battery": 50,
            "temperature": 70,
            "gravity": 1.026,
            "created": "2017-07-13T09:25:37.164272351Z"
          },
          {
            "power": -67,
            "battery": 50,
            "temperature": 71,
            "gravity": 1.027,
            "created": "2017-07-13T08:25:37.164272351Z"
          },
          {
            "power": -62,
            "battery": 50,
            "temperature": 70,
            "gravity": 1.029,
            "created": "2017-07-13T07:25:37.164272351Z"
          },
          {
            "power": -63,
            "battery": 50,
            "temperature": 70,
            "gravity": 1.031,
            "created": "2017-07-13T06:25:37.164272351Z"
          },
          {
            "power": -61,
            "battery": 50,
            "temperature": 70,
            "gravity": 1.033,
            "created": "2017-07-13T05:25:37.164272351Z"
          },
          {
            "power": -55,
            "battery": 50,
            "temperature": 71,
            "gravity": 1.037,
            "created": "2017-07-13T04:25:37.164272351Z"
          },
          {
            "power": -59,
            "battery": 50,
            "temperature": 70,
            "gravity": 1.040,
            "created": "2017-07-13T03:25:37.164272351Z"
          },
          {
            "power": -65,
            "battery": 50,
            "temperature": 70,
            "gravity": 1.044,
            "created": "2017-07-13T02:25:37.164272351Z"
          },
          {
            "power": -59,
            "battery": 50,
            "temperature": 70,
            "gravity": 1.048,
            "created": "2017-07-13T01:25:37.164272351Z"
          },
          {
            "power": -62,
            "battery": 50,
            "temperature": 70,
            "gravity": 1.053,
            "created": "2017-07-13T00:25:37.164272351Z"
          }
        ]
      }
    },
    mounted() {
      this.loadDevices()

      setInterval(function () {
        this.loadDevices()
        console.log("Reloading devices...")
      }.bind(this), 30000)
    },
    methods: {
      loadDevices() {
        this.$http.get('/api/v1/devices').then(response => {
          this.devices = response.body
        }, response => {
          console.log('ERROR')
          console.log(response)
        })
      },
      dataForDeviceGravity(device) {
        if (device === undefined || device === null) {
          return {labels: {}, datasets: []}
        }

        var labels = null
        var values = null
        this.$http.get('/api/v1/devices/' + device.id + '/metrics').then(response => {
          labels = response.body.map(function(m) {
            return new Date(m.created).toLocaleDateString()
          })
          values = response.body.map(function(m) {
            return m.gravity
          })
        }, response => {
          console.log("Error retrieving metrics")
          console.log(response)
        })
        return {labels: labels, datasets: [{label: "Specific Gravity", data: values}]}
      },
      showGraphs(device) {
        console.log("Showing graphs for device: " + device.id)
        this.deviceMetrics = device
        $('#metricsModal').modal('open')
      },
      showSettings(device) {
        console.log("Showing settings for device: " + device.id)
        this.deviceSettings = device
        $('#settingsModal').modal('open')
      },
      updateDevice(device) {
        this.$http.post('/api/v1/devices/' + device.id, device).then(response => {
          console.log('SUCCESS')
          console.log(response)
          this.loadDevices()
        }, response => {
          console.log('ERROR')
          console.log(response)
        })
        $('#settingsModal').modal('close')
      },
      refreshDevice(deviceID) {
        console.log("Refreshing device: " + deviceID)
      },
      resetDevice(deviceID) {
        console.log("Resetting device: " + deviceID)
      },
      deleteDevice(deviceID) {
        console.log("Deleting device: " + deviceID)
      },
      capCase (word) {
        return word.replace(/_/g, ' ').split(' ').map((w) => {
          return w.charAt(0).toUpperCase() + w.slice(1)
        }).join(' ')
      },
      timeAgo (t) {
        return moment(t).fromNow()
      },
      getBatteryIcon (perc) {
        if (perc < 100) {
          var icon = 'mdi-battery-' + perc
          return [icon]
        }
        return ['mdi-battery']
      }
    },
    watch: {
        devices (newDevices) {
          for (var i=0; i<newDevices.length; i++) {
            var device = newDevices[i]
            for (var c=0; c<this.devices.length; c++) {
              if (device.id === this.devices[c].id) {
                device.latest_metrics.gravityChangeInt = device.latest_metrics.gravity - this.devices[c].latest_metrics.gravity
                device.latest_metrics.gravityChangePerc = (device.latest_metrics.gravity / this.devices[c].latest_metrics.gravity) * 100
              }
            }
          }
        }
    },
  }
</script>

<style>
    @font-face {
        font-family: 'Montserrat';
        font-style: normal;
        font-weight: 400;
        src: local('Montserrat Regular'), local('Montserrat-Regular'), url(/static/fonts/montserrat.woff2) format('woff2');
        unicode-range: U+0000-00FF, U+0131, U+0152-0153, U+02C6, U+02DA, U+02DC, U+2000-206F, U+2074, U+20AC, U+2212, U+2215;
    }

    body {
        background: #544947;
        font-family: Montserrat, Arial, sans-serif;
    }

    h2 {
        font-size: 14px;
    }

    .widget {
        position: relative;
        width: 100%;
        margin: 0px auto;
        padding: 0px 10px;
        height: 80px;
        color: #FFF;
    }

    .name-wrapper {
        max-width: 25%;
        width: 25%;
        padding: 10px 0px 0px 0px;
    }

    .name {
        text-align: left;
        font-size: 24px;
    }

    .name:hover {
        text-decoration-line: underline;
        text-decoration-style: dashed;
    }

    .timestamp {
        font-size: 14px;
        color: #DBDBDB;
        text-align: left;
    }

    .gravity-wrapper {
        position: relative;
        max-width: 23%;
        width: 23%;
    }

    .gravity {
        font-size: 36px;
        position: absolute;
        right: 10px;
        top: 2px;
    }

    .gravity-stats {
        font-size: 16px;
        position: absolute;
        right: 10px;
        top: 48px;
    }

    .temperature,
    .battery,
    .signal {
        max-width: 15%;
        width: 15%;
    }

    .settings {
        top: 18px;
        right: 15px;
        font-size: 32px;
        position: absolute;
    }

    .settings a {
        color: #fff;
    }

    .widget .warning {
        top: 65px;
        right: 10px;
        font-size: 30px;
        position: absolute;
    }

    .info {
        float: left;
        height: 100%;
        text-align: center;
        color: #fff;
        padding-top: 10px;
    }

    .info span {
        display: inline-block;
        font-size: 30px;
        margin-top: 10px;
    }

    .warning {
        color: #ff9800;
    }

    .black .widget {
        background-color: #000;
    }

    .red .widget {
        background-color: #F44336;
    }

    .green .widget {
        background-color: #4CAF50;
    }

    .purple .widget {
        background-color: #9c27b0;
    }

    .orange .widget {
        background-color: #ff9800;
    }

    .blue .widget {
        background-color: #2196F3;
    }

    .pink .widget {
        background-color: #e91e63;
    }
</style>
