<script lang="ts" setup>
  import { ref, watch } from 'vue'
  import { getTemplateVersionsNameByIds } from '../../../../../../../api/template';
  import { updateBoundTemplateVersion } from '../../../../../../../api/config';
  import { ITemplateVersionsName } from '../../../../../../../../types/template';
import { Message } from 'bkui-vue';

  interface ITplVersionItem {
    id: number;
    name: string;
    isLatest: boolean;
  }

  const props = defineProps<{
    show: boolean;
    bkBizId: string;
    appId: number;
    pkgId: number;
    bindingId: number;
    templateId: number;
    versionId: number;
    versionName: string;
  }>()

  const emits = defineEmits(['update:show', 'updated'])

  const loading = ref(false)
  const versionList = ref<ITplVersionItem[]>([])
  const selected = ref()
  const formRef = ref()
  const pending = ref(false)

  watch(() => props.show, val => {
    if (val) {
      selected.value = undefined
      getTemplateVersions()
    }
  })

  const getTemplateVersions = async () => {
    loading.value = true
    const res = await getTemplateVersionsNameByIds(props.bkBizId, [props.templateId])
    const templateVersion: ITemplateVersionsName = res.details[0]
    const list: ITplVersionItem[] = []
    templateVersion.template_revisions.forEach(item => {
      const { template_revision_id, template_revision_name } = item
      list.push({ id: template_revision_id, name: template_revision_name, isLatest: false })
    })
    const latestVersion = templateVersion.template_revisions.find(version => version.template_revision_id === templateVersion.latest_template_revision_id)
    if (latestVersion) {
      list.unshift({
        id: latestVersion.template_revision_id,
        name: `latest（当前最新为${latestVersion.template_revision_name}）`,
        isLatest: true
      })
      selected.value = 0
    }
    versionList.value = list
    loading.value = false
  }

  const handleReplaceConfirm = async() => {
    await formRef.value.validate()
    const isLatest = selected.value === 0
    let versionId = selected.value
    if (isLatest) {
      const id = versionList.value.find(item => item.isLatest)?.id
      versionId = id
    }
    const params = {
      bindings: [{
        template_set_id: props.pkgId,
        template_revisions: [{
          template_id: props.templateId,
          template_revision_id: versionId,
          is_latest: isLatest
        }]
      }]
    }
    await updateBoundTemplateVersion(props.bkBizId, props.appId, props.bindingId, params)
    emits('updated')
    close()
    Message({
      theme: 'success',
      message: '模板版本更新成功'
    })
  }

  const close = () => {
    emits('update:show')
  }

</script>
<template>
  <bk-dialog
    title="替换版本"
    head-align="left"
    footer-align="right"
    width="480"
    :is-show="props.show"
    :is-loading="loading || pending"
    @confirm="handleReplaceConfirm"
    @closed="close">
    <bk-form ref="formRef" :model="{ selected }">
      <bk-form-item label="当前版本">
        <div>{{ props.versionName }}</div>
      </bk-form-item>
      <bk-form-item label="目标版本" required property="selected">
        <bk-select v-model="selected" :loading="loading" :clearable="false" :filterable="true" :input-search="false">
          <bk-option
            v-for="option in versionList"
            v-overflow-title
            :key="option.isLatest ? 0 : option.id"
            :value="option.isLatest ? 0 : option.id"
            :label="option.name">
          </bk-option>
        </bk-select>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>
