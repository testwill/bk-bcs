<script lang="ts" setup>
  import { ref, computed, watch } from 'vue'
  import { IPackagesCitedByApps } from '../../../../../../types/template';
  import { getUnNamedVersionAppsBoundByLatestTemplateVersion } from '../../../../../api/template'
  import LinkToApp from '../../list/components/link-to-app.vue'

  const props = defineProps<{
    show: boolean;
    spaceId: string;
    templateSpaceId: number;
    templateId: number;
    versionId: number;
    pending: boolean;
  }>()

  const emits = defineEmits(['update:show', 'confirm'])

  const loading = ref(false)
  const citedList = ref<IPackagesCitedByApps[]>([])

  const maxTableHeight = computed(() => {
    const windowHeight = window.innerHeight
    return windowHeight * 0.6 - 200
  })

  watch(() => props.show, val => {
    if (val) {
      getCitedData()
    }
  })

  const getCitedData = async() => {
    loading.value = true
    const params = {
      start: 0,
      all: true
    }
    const res = await getUnNamedVersionAppsBoundByLatestTemplateVersion(props.spaceId, props.templateSpaceId, props.templateId, params)
    citedList.value = res.details
    loading.value = false
  }


  const close = () => {
    emits('update:show', false)
  }

  defineExpose({
    close
  })
</script>
<template>
  <bk-dialog
    ext-cls="create-version-confirm-dialog"
    title="确认更新配置项版本？"
    header-align="center"
    footer-align="center"
    :width="400"
    :is-show="props.show"
    :esc-close="false"
    :quick-close="false"
    @closed="close">
    <p class="tips">以下套餐及服务未命名版本中引用的此配置项也将更新</p>
    <bk-loading style="min-height: 100px;" :loading="loading">
      <bk-table :data="citedList" :max-height="maxTableHeight">
        <bk-table-column label="所在套餐" prop="template_set_name"></bk-table-column>
        <bk-table-column label="引用此模板的服务">
          <template #default="{ row }">
            <div v-if="row.app_id" class="app-info">
              <div v-overflow-title class="name-text">{{ row.app_name }}</div>
              <LinkToApp class="link-icon" :id="row.app_id" />
            </div>
          </template>
        </bk-table-column>
      </bk-table>
    </bk-loading>
    <template #footer>
      <div class="actions-wrapper">
        <bk-button theme="primary" :loading="pending" @click="emits('confirm')">确定</bk-button>
        <bk-button @click="close">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>
<style lang="scss" scoped>
  .app-info {
    display: flex;
    align-items: center;
    overflow: hidden;
    .name-text {
      overflow: hidden;
      white-space: nowrap;
      text-overflow: ellipsis;
    }
    .link-icon {
      flex-shrink: 0;
      margin-left: 10px;
    }
  }
  .actions-wrapper {
    padding-bottom: 20px;
    .bk-button:not(:last-of-type) {
      margin-right: 8px;
    }
  }
</style>
<style lang="scss">
  .create-version-confirm-dialog.bk-modal-wrapper.bk-dialog-wrapper {
    .bk-modal-footer {
      padding: 32px 0 48px;
      background: #ffffff;
      border-top: none;
      .bk-button {
        min-width: 88px;
      }
    }
  }
</style>
