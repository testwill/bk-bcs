<script lang="ts" setup>
  import { ref, watch } from 'vue'
  import { storeToRefs } from 'pinia'
  import { Message } from 'bkui-vue/lib'
  import { useGlobalStore } from '../../../../../store/global'
  import { updateTemplateSpace } from '../../../../../api/template'

  const { spaceId } = storeToRefs(useGlobalStore())

  const props = defineProps<{
    show: boolean,
    data: { id: number; name: string; memo: string }
  }>()

  const emits = defineEmits(['update:show', 'edited'])

  const isShow = ref(false)
  const formRef = ref()
  const pending = ref(false)
  const memo = ref('')

  watch(() => props.show, val => {
    isShow.value = val
    if (val) {
      memo.value = props.data.memo
    }
  })

  const handleEdit = () => {
    formRef.value.validate().then(async () => {
      try {
        pending.value = true
        await updateTemplateSpace(spaceId.value, props.data.id, { memo: memo.value })
        handleClose()
        emits('edited')
        Message({
          theme: 'success',
          message: '编辑成功'
        })
      } catch (e) {
        console.error(e)
      } finally {
        pending.value = false
      }
    })
  }

  const handleClose = () => {
    emits('update:show', false)
  }

</script>
<template>
  <bk-dialog
    ext-cls="edit-template-space-dialog"
    confirm-text="保存"
    :width="480"
    :is-show="isShow"
    :is-loading="pending"
    :esc-close="false"
    :quick-close="false"
    @confirm="handleEdit"
    @closed="handleClose">
    <template #header>
      <div class="header-wrapper">
        <div class="title">编辑空间</div>
        <div class="space-name">{{ props.data.name }}</div>
      </div>
    </template>
    <bk-form ref="formRef" form-type="vertical" :model="{ memo }">
      <bk-form-item label="模板空间描述" property="memo">
        <bk-input v-model="memo" type="textarea" :rows="6" :maxlength="100" />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>
<style lang="postcss" scoped>
  .header-wrapper {
    display: flex;
    align-items: center;
    .title {
      margin-right: 16px;
      padding-right: 16px;
      line-height: 24px;
      border-right: 1px solid #dcdee5;
    }
    .space-name {
      flex: 1;
      line-height: 24px;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
  }
</style>
