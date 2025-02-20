<script lang="ts" setup>
  import { ref, watch } from 'vue'
  import { Message } from 'bkui-vue';
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../../types/config'
  import { createServiceConfigItem, updateConfigContent } from '../../../../../../../../api/config'
  import { getConfigEditParams } from '../../../../../../../../utils/config'
  import useModalCloseConfirmation from '../../../../../../../../utils/hooks/use-modal-close-confirmation'
  import ConfigForm from '../config-form.vue'

  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
  }>()

  const emits = defineEmits(['update:show', 'confirm'])

  const configForm = ref<IConfigEditParams>(getConfigEditParams())
  const fileUploading = ref(false)
  const pending = ref(false)
  const content = ref<IFileConfigContentSummary|string>('')
  const formRef = ref()
  const isFormChange = ref(false)

  watch(() => props.show, val => {
    console.log(val)
    if (val) {
      configForm.value = getConfigEditParams()
    }
  })

  const handleFormChange = (data: IConfigEditParams, configContent: IFileConfigContentSummary|string) => {
    configForm.value = data
    content.value = configContent
  }

  const handleBeforeClose = async () => {
    if (isFormChange.value) {
      const result = await useModalCloseConfirmation()
      return result
    }
    return true
  }

  const handleSubmit = async() => {
    const isValid = await formRef.value.validate()
    if (!isValid) return

    try {
      pending.value = true
      let sign = await formRef.value.getSignature()
      let size = 0
      if (configForm.value.file_type === 'binary') {
        size = Number((<IFileConfigContentSummary>content.value).size)
      } else {
        const stringContent = <string>content.value
        size = new Blob([stringContent]).size
        await updateConfigContent(props.bkBizId, props.appId, stringContent, sign)
      }
      const params = { ...configForm.value, ...{ sign, byte_size: size } }
      await createServiceConfigItem(props.appId, props.bkBizId, params)
      emits('confirm')
      close()
      Message({
        theme: 'success',
        message: '新建配置项成功'
      })
    }catch (e) {
      console.log(e)
    } finally {
      pending.value = false
    }
  }

  const close = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-sideslider
    width="640"
    title="新增配置文件"
    :is-show="props.show"
    :before-close="handleBeforeClose"
    @closed="close">
    <ConfigForm
      ref="formRef"
      class="config-form-wrapper"
      v-model:fileUploading="fileUploading"
      :config="configForm"
      :content="content"
      :editable="true"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      @change="handleFormChange"/>
    <section class="action-btns">
      <bk-button theme="primary" :loading="pending" :disabled="fileUploading" @click="handleSubmit">保存</bk-button>
      <bk-button @click="close">取消</bk-button>
    </section>
  </bk-sideslider>
</template>
<style lang="scss" scoped>
  .config-form-wrapper {
    padding: 20px 40px;
    height: calc(100vh - 101px);
    overflow: auto;
  }
  .action-btns {
    border-top: 1px solid #dcdee5;
    padding: 8px 24px;
    .bk-button {
      margin-right: 8px;
      min-width: 88px;
    }
  }
</style>
