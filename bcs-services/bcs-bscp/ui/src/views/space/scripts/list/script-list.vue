<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { useRouter } from 'vue-router'
  import { InfoBox } from 'bkui-vue'
  import { Plus, Search } from 'bkui-vue/lib/icon'
  import { storeToRefs } from 'pinia'
  import { useGlobalStore } from '../../../../store/global'
  import { useScriptStore } from '../../../../store/script'
  import { getScriptList, getScriptTagList, deleteScript } from '../../../../api/script'
  import { IScriptItem, IScriptTagItem, IScriptListQuery} from '../../../../../types/script'
  import { datetimeFormat } from '../../../../utils/index'
  import CreateScript from './create-script.vue'
  import ScriptCited from './script-cited.vue'

  const { spaceId } = storeToRefs(useGlobalStore())
  const {versionListPageShouldOpenEdit } = storeToRefs(useScriptStore())
  const router = useRouter()

  const showCreateScript = ref(false)
  const showCiteSlider = ref(false)
  const scriptsData = ref<IScriptItem[]>([])
  const scriptsLoading = ref(false)
  const tagsData = ref<IScriptTagItem[]>([])
  const tagsLoading = ref(false)
  const showAllTag = ref(true) // 全部脚本
  const selectedTag = ref('') // 未分类或具体tag下脚本
  const currentId = ref(0)
  const searchStr = ref('')
  const pagination = ref({
    current: 1,
    count: 0,
    limit: 10,
  })

  watch(() => spaceId.value, () => {
    refreshList()
    getTags()
  })

  onMounted(() => {
    getScripts()
    getTags()
  })

  // 获取脚本列表
  const getScripts = async () => {
    scriptsLoading.value = true
    const params: IScriptListQuery = {
      start: (pagination.value.current - 1) * pagination.value.limit,
      limit: pagination.value.limit,
    }
    if (showAllTag.value) {
      params.all = true
    } else {
      if (selectedTag.value === '') {
        params.not_tag = true
      } else {
        params.tag = selectedTag.value
      }
    }
    if (searchStr.value) {
      params.name = searchStr.value
    }
    const res = await getScriptList(spaceId.value, params)
    scriptsData.value = res.details
    pagination.value.count = res.count
    scriptsLoading.value = false
  }

  // 获取标签列表
  const getTags = async () => {
    tagsLoading.value = true
    const res = await getScriptTagList(spaceId.value)
    tagsData.value = res.details
    tagsLoading.value = false
  }

  const handleSelectTag = (tag: string, all: boolean = false) => {
    searchStr.value = ''
    selectedTag.value = tag
    showAllTag.value = all
    refreshList()
  }

  const handleOpenCitedSlider = (id: number) => {
    currentId.value = id
    showCiteSlider.value = true
  }

  const handleEditClick = (script: IScriptItem) => {
    router.push({ name: 'script-version-manage', params: { spaceId: spaceId.value, scriptId: script.hook.id } })
    versionListPageShouldOpenEdit.value = true
  }

  // 删除分组
  const handleDeleteScript = (script: IScriptItem) => {
    InfoBox({
      title: `确认是否删除脚本【${script.hook.spec.name}?】`,
      subTitle: `${script.confirm_delete ? '当前脚本有被服务未命名版本引用，删除后，未命名版本里的引用将会被删除，是否确认删除？' : ''}`,
      infoType: "warning",
      headerAlign: "center" as const,
      footerAlign: "center" as const,
      onConfirm: async () => {
        await deleteScript(spaceId.value, script.hook.id)
        if (scriptsData.value.length === 1 && pagination.value.current > 1) {
          pagination.value.current = pagination.value.current - 1
        }
        getScripts()
      },
    } as any)
  }

  const handleNameInputChange = (val: string) => {
    if (!val) {
      refreshList()
    }
  }

  const handleCreatedScript = () => {
    refreshList()
    getTags()
  }

  const refreshList = () => {
    pagination.value.current = 1
    getScripts()
  }

  const handlePageLimitChange = (val: number) => {
    pagination.value.limit = val
    refreshList()
  }
</script>
<template>
  <section class="scripts-manange-page">
    <div class="side-menu">
      <div class="group-wrapper">
        <li :class="['group-item', { actived: showAllTag }]" @click="handleSelectTag('', true)">
          <i class="bk-bscp-icon icon-block-shape group-icon"></i>全部脚本
        </li>
        <li :class="['group-item', { actived: !showAllTag && selectedTag === '' }]"  @click="handleSelectTag('')">
          <i class="bk-bscp-icon icon-tags group-icon"></i>未分类
        </li>
      </div>
      <div class="custom-tag-list">
        <li
          v-for="(item, index) in tagsData" :key="index"
          :class="['group-item', { actived: selectedTag === item.tag }]"
          @click="handleSelectTag(item.tag)">
          <i class="bk-bscp-icon icon-tags group-icon"></i>
          <span class="name">{{ item.tag }}</span>
          <span class="num">{{ item.counts }}</span>
        </li>
      </div>
    </div>
    <div class="script-list-wrapper">
      <div class="operate-area">
        <bk-button theme="primary" @click="showCreateScript = true"><Plus class="button-icon" />新建脚本</bk-button>
        <bk-input
          v-model="searchStr"
          class="search-script-input"
          placeholder="脚本名称"
          :clearable="true"
          @enter="refreshList"
          @clear="refreshList"
          @change="handleNameInputChange">
           <template #suffix>
              <Search class="search-input-icon" />
           </template>
        </bk-input>
      </div>
      <bk-loading style="min-height: 300px;" :loading="scriptsLoading">
        <bk-table
          :border="['outer']"
          :data="scriptsData"
          :pagination="pagination"
          @page-limit-change="handlePageLimitChange"
          @page-change="refreshList">
          <bk-table-column label="脚本名称">
            <template #default="{ row }">
              <div v-if="row.hook" class="hook-name" @click="router.push({ name: 'script-version-manage', params: { spaceId, scriptId: row.hook.id } })">{{ row.hook.spec.name }}</div>
            </template>
          </bk-table-column>
          <bk-table-column label="脚本语言" prop="hook.spec.type" width="120"></bk-table-column>
          <bk-table-column label="分类标签">
            <template #default="{ row }">
              <span v-if="row.hook">{{ row.hook.spec.tag || '--' }}</span>
            </template>
          </bk-table-column>
          <bk-table-column label="被引用" width="100">
          <template #default="{ row }">
              <bk-button v-if="row.bound_num > 0" text theme="primary" @click="handleOpenCitedSlider(row.hook.id)">{{ row.bound_num }}</bk-button>
              <span v-else>0</span>
          </template>
        </bk-table-column>
        <bk-table-column label="更新人" prop="hook.revision.reviser"></bk-table-column>
        <bk-table-column label="更新时间" width="220">
          <template #default="{ row }">
            <span v-if="row.hook">{{ datetimeFormat(row.hook.revision.update_at) }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="操作">
          <template #default="{ row }" width="180">
            <div class="action-btns">
              <!-- <bk-button text theme="primary" @click="handleEditClick(row)">编辑</bk-button> -->
              <bk-button text theme="primary" @click="router.push({ name: 'script-version-manage', params: { spaceId, scriptId: row.hook.id } })">版本管理</bk-button>
              <bk-button text theme="primary" @click="handleDeleteScript(row)">删除</bk-button>
            </div>
          </template>
        </bk-table-column>
        </bk-table>
      </bk-loading>
    </div>
    <CreateScript v-if="showCreateScript" v-model:show="showCreateScript" @created="handleCreatedScript" />
    <ScriptCited v-model:show="showCiteSlider" :id="currentId" />
  </section>
</template>
<style lang="scss" scoped>
  .scripts-manange-page {
    display: flex;
    align-items: center;
    height: 100%;
    background: #ffffff;
  }
  .side-menu {
    padding: 16px 0;
    width: 280px;
    height: 100%;
    background: #f5f7fa;
    box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.15);
    z-index: 1;
    .group-wrapper {
      padding-bottom: 16px;
      border-bottom: 1px solid #dcdee5;
    }
    .group-item {
      position: relative;
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 8px 16px 8px 44px;
      color: #313238;
      font-size: 12px;
      cursor: pointer;
      &:hover {
        color: #348aff;
        .group-icon {
          color: #3a84ff;
        }
      }
      &.actived {
        background: #e1ecff;
        color: #3a84ff;
        .group-icon {
          color: #3a84ff;
        }
        .num {
          color: #ffffff;
          background: #a3c5fd;
        }
      }
      .group-icon {
        position: absolute;
        top: 9px;
        left: 22px;
        margin-right: 8px;
        font-size: 14px;
        color: #979ba5;
      }
      .name {
        flex: 0 1 auto;
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
      }
      .num {
        flex: 0 0 auto;
        padding: 0 8px;
        height: 16px;
        line-height: 16px;
        color: #979ba5;
        background: #f0f1f5;
        border-radius: 2px;
      }
    }
    .custom-tag-list {
      padding: 16px 0;
      height: calc(100% - 82px);
      overflow: auto;
    }
  }
  .script-list-wrapper {
    padding: 24px;
    width: calc(100% - 280px);
    height: 100%;
    background: #ffffff;
  }
  .operate-area {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    .button-icon {
      font-size: 18px;
    }
  }
  .search-script-input {
    width: 320px;
  }
  .search-input-icon {
    padding-right: 10px;
    color: #979ba5;
    background: #ffffff;
  }
  .action-btns {
    .bk-button {
      margin-right: 8px;
    }
  }
  .hook-name {
    color: #348aff;
    cursor: pointer;
  }
</style>
