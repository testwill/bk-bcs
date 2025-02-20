<script setup lang="ts">
  import { ref, watch, computed, onMounted } from 'vue'
  import { storeToRefs } from 'pinia'
  import { DownShape } from 'bkui-vue/lib/icon'
  import { useConfigStore } from '../../../../../../../store/config'
  import { IConfigItem, IConfigListQueryParams, IBoundTemplateDetail } from '../../../../../../../../types/config'
  import { ICommonQuery } from '../../../../../../../../types/index'
  import { getConfigList, getBoundTemplates } from '../../../../../../../api/config'
  import { getConfigTypeName } from '../../../../../../../utils/config'
  import SearchInput from '../../../../../../../components/search-input.vue'
  import EditConfig from '../config-table-list/edit-config.vue'
  import ViewConfig from '../config-table-list/view-config.vue'

  interface IConfigsGroupData {
    id: number;
    name: string;
    expand: boolean;
    configs: IConfigTableItem[];
  }

  interface IConfigTableItem {
    id: number;
    name: string;
    file_type: string;
    versionId: number;
    versionName: string;
    path: string;
    creator: string;
    reviser: string;
    update_at: string;
    file_state: string;
  }

  const store = useConfigStore()
  const { versionData } = storeToRefs(store)

  const props = defineProps<{
    bkBizId: string,
    appId: number,
  }>()

  const loading = ref(false)
  const commonConfigListLoading = ref(false)
  const configList = ref<IConfigItem[]>([]) // 非模板配置项
  const boundTemplateListLoading = ref(false)
  const templateList = ref<IBoundTemplateDetail[]>([]) // 配置项模板
  const tableGroupsData = ref<IConfigsGroupData[]>([])
  const searchStr = ref('')
  const editConfigSliderData = ref({
    open: false,
    id: 0
  })
  const viewConfigSliderData = ref({
    open: false,
    data: { id: 0, type: '' }
  })

  watch(() => versionData.value.id, () => {
    getListData()
  })

  // 是否为未命名版本
  const isUnNamedVersion = computed(() => {
    return versionData.value.id === 0
  })

  onMounted(() => {
    getListData()
  })

  const getListData = async () => {
    // 拉取到版本列表之前不加在列表数据
    if (typeof versionData.value.id !== 'number') {
      return
    }
    loading.value = true
    await Promise.all([
      getCommonConfigList(),
      getBoundTemplateList()
    ])
    tableGroupsData.value = transListToTableData()
    loading.value = false
  }


  // 获取非模板配置项列表
  const getCommonConfigList = async() => {
    commonConfigListLoading.value = true
    try {
      const params: IConfigListQueryParams = {
        start: 0,
        all: true
      }
      if (searchStr.value) {
        params.searchKey = searchStr.value
      }
      if (!isUnNamedVersion.value) {
        params.release_id = versionData.value.id
      }
      const res = await getConfigList(props.bkBizId, props.appId, params)
      configList.value = res.details
    } catch (e) {
      console.error(e)
    } finally {
      commonConfigListLoading.value = false
    }
  }

  // 获取模板配置项列表
  const getBoundTemplateList = async() => {
    boundTemplateListLoading.value = true
    try {
      const params: ICommonQuery = {
        start: 0,
        all: true
      }
      if (searchStr.value) {
        params.search_fields = 'revision_name,revision_memo,name,path,creator'
        params.search_value = searchStr.value
      }
      const res = await getBoundTemplates(props.bkBizId, props.appId, params)
      templateList.value = res.details
    } catch (e) {
      console.error(e)
    } finally {
      boundTemplateListLoading.value = false
    }
  }

  const transListToTableData = () => {
    const pkgsGroups = groupTplsByPkg(templateList.value)
    return [
      { id: 0, name: '非模板配置', expand: true, configs: transConfigsToTableItemData(configList.value) },
      ...pkgsGroups
    ]
  }

  // 将非模板配置项数据转为表格数据
  const transConfigsToTableItemData = (list: IConfigItem[]) => {
    return list.map((item: IConfigItem) => {
        const { id, spec, revision, file_state } = item
        const { name, file_type, path } = spec
        const { creator, reviser, update_at } = revision
        return { id, name, versionId: 0, versionName: '--', path, creator, reviser, update_at, file_type, file_state }
      })
  }

  // 将模板按套餐分组，并将模板数据格式转为表格数据
  const groupTplsByPkg = (list: IBoundTemplateDetail[]) => {
    const groups: IConfigsGroupData[] = []
    list.forEach(tpl => {
      const {
        template_space_name, template_set_id, template_set_name,
        template_id: id, template_revision_id: versionId,
        template_revision_name: versionName, name, file_type, path, creator
      } = tpl
      const config = { id, name, file_type, versionId, versionName, path, creator, reviser: '--', update_at: '--', file_state: '' }
      const group = groups.find(item => item.id === template_set_id)
      if (group) {
        group.configs.push(config)
      } else {
        groups.push({
          id: template_set_id,
          name: `${template_space_name} - ${template_set_name}`,
          expand: false,
          configs: [config]
        })
      }
    })
    return groups
  }

  const handleConfigClick = (config: IConfigTableItem, groupId: number) => {
    if (isUnNamedVersion.value && groupId === 0) {
      editConfigSliderData.value = {
        open: true,
        id: config.id
      }
    } else {
      const id = groupId === 0 ? config.id : config.versionId
      const type = groupId === 0 ? 'config' : 'template'
      viewConfigSliderData.value = {
        open: true,
        data: { id, type }
      }
    }

  }

</script>
<template>
  <section class="config-list-with-templates">
    <SearchInput
      v-model="searchStr"
      class="config-search-input"
      placeholder="配置文件名/创建人/修改人"
      @search="getListData" />
    <bk-loading class="loading-wrapper" :loading="loading">
      <div v-for="group in tableGroupsData" :key="group.id" class="config-group">
        <div class="group-title" @click="group.expand = !group.expand">
          <DownShape :class="['fold-icon', { fold: !group.expand }]" />
          {{ group.name }}
        </div>
        <div v-if="group.expand" class="config-list-wrapper">
          <div
            v-for="config in group.configs"
            :class="['config-item', { disabled: config.file_state === 'DELETE' }]"
            :key="config.id"
            @click="handleConfigClick(config, group.id)">
            <div class="config-name">{{ config.name }}</div>
            <div class="config-type">{{ getConfigTypeName(config.file_type) }}</div>
          </div>
          <bk-exception v-if="group.configs.length === 0" scene="part" type="empty" description="暂无配置项"></bk-exception>
        </div>
      </div>
    </bk-loading>
    <EditConfig
      v-model:show="editConfigSliderData.open"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId"
      :config-id="editConfigSliderData.id" />
    <ViewConfig
      v-model:show="viewConfigSliderData.open"
      v-bind="viewConfigSliderData.data"
      :bk-biz-id="props.bkBizId"
      :app-id="props.appId" />
  </section>
</template>
<style lang="scss" scoped>
  .config-list-with-templates {
    padding: 24px;
    height: 100%;
    background: #fafbfd;
    overflow: auto;
  }
  .config-search-input {
    margin-bottom: 16px;
  }
  .loading-wrapper {
    height: calc(100% - 48px);
    overflow: auto;
  }
  .group-title {
    display: flex;
    align-items: center;
    margin: 8px 0;
    line-height: 20px;
    font-size: 12px;
    color: #63656e;
    cursor: pointer;
    .fold-icon {
      margin-right: 8px;
      font-size: 14px;
      color: #3a84ff;
      transition: transform 0.2s ease-in-out;
      &.fold {
        color: #c4c6cc;
        transform: rotate(-90deg);
      }
    }
  }
  .config-list-wrapper {
    max-height: 472px; // 每个分组最多显示10条，超出后滚动显示
    overflow: auto;
  }
  .config-item {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
    font-size: 12px;
    background: #ffffff;
    box-shadow: 0 1px 1px 0 rgba(0, 0, 0, 0.06);
    border-radius: 2px;
    cursor: pointer;
    &.disabled {
      cursor: not-allowed;
      .config-type,
      .config-name {
        color: #dcdee5;
      }
    }
    &:not(.disabled):hover {
      background: #e1ecff;
    }
    .config-name {
      padding: 0 16px;
      width: 242px;
      height: 40px;
      line-height: 40px;
      color: #313238;
      white-space: nowrap;
      text-overflow: ellipsis;
      overflow: hidden;
    }
    .config-type {
      color: #979ba5;
    }
  }
</style>
