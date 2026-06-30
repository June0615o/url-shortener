<script setup>
import { ref } from 'vue'
import QRCode from 'qrcode'

const props = defineProps({ url: String })
const visible = ref(false)
const qrDataUrl = ref('')

async function open() {
  visible.value = true
  qrDataUrl.value = await QRCode.toDataURL(props.url, { width: 256, margin: 2 })
}
</script>

<template>
  <el-button size="small" @click="open">📱 二维码</el-button>
  <el-dialog v-model="visible" title="扫码访问" width="320px" center>
    <div style="text-align:center">
      <img v-if="qrDataUrl" :src="qrDataUrl" alt="QR Code" style="width:256px;height:256px" />
      <p style="margin-top:12px;word-break:break-all;font-size:12px;color:#909399">{{ url }}</p>
    </div>
  </el-dialog>
</template>
