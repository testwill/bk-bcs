<script lang="ts" setup>
  import { ref } from 'vue'
  import { ITemplateConfigItem } from '../../../../../../../../types/template';
  import useDeleteTemplateConfigs from '../../../../../../../utils/hooks/use-delete-template-configs';
  import DeleteConfigDialog from './delete-config-dialog.vue';


  const props = defineProps<{
    spaceId: string;
    currentTemplateSpace: number;
    configs: ITemplateConfigItem[];
  }>()

  const emits = defineEmits(['deleted'])

  const isDialogShow = ref(false)

  // const pending = ref(false)

  // const handleDeleteTemplate = async () => {
  //   try {
  //     pending.value = true
  //     const result = await useDeleteTemplateConfigs(props.spaceId, props.currentTemplateSpace, props.configs)
  //     if (result) {
  //       emits('deleted')
  //     }
  //   } catch (e) {
  //     console.log(e)
  //   } finally {
  //     pending.value = false
  //   }
  // }

</script>
<template>
  <bk-button :disabled="props.configs.length === 0"  @click="isDialogShow = true">批量删除</bk-button>
  <DeleteConfigDialog v-model:show="isDialogShow" :configs="props.configs" @deleted="emits('deleted')" />
</template>
