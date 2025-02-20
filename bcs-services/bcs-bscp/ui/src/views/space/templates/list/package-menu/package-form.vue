<script lang="ts" setup>
  import { onMounted, ref, watch, computed } from 'vue'
  import { cloneDeep } from 'lodash'
  import { storeToRefs } from 'pinia'
  import { useUserStore } from '../../../../../store/user'
  import { ITemplatePackageEditParams } from '../../../../../../types/template'
  import { getAppList } from '../../../../../api/index'
  import { IAppItem } from '../../../../../../types/app'

  const { userInfo } = storeToRefs(useUserStore())

  const props = defineProps<{
    spaceId: string;
    data: ITemplatePackageEditParams,
    apps?: number[]; // 套餐绑定的服务，编辑时需要区分哪些服务被去掉
  }>()

  const emits = defineEmits(['change'])

  const localVal = ref<ITemplatePackageEditParams>({
    name: '',
    memo: '',
    public: true,
    bound_apps: []
  })
  const formRef = ref()
  const serviceLoading = ref(false)
  const serviceList = ref<IAppItem[]>([])
  const rules = {
    public: [
      {
        validator: (val: boolean) => {
          if (!val && localVal.value.bound_apps.length === 0) {
            return false
          }
          return true
        },
        message: '指定服务不能为空'
      }
    ]
  }

  watch(() => props.data, (val) => {
    localVal.value = cloneDeep(val)
  }, {
    immediate: true
  })

  const deletedApps = computed(() => {
    const changed: IAppItem[] = []
    if (props.apps) {
      props.apps.forEach(id => {
        if (!localVal.value.bound_apps.includes(id)) {
          const app = serviceList.value.find(item => item.id === id)
          if (app) {
            changed.push(app)
          }
        }
      })
    }
    return changed
  })

  onMounted(() => {
    getServiceList()
  })

  const getServiceList = async () => {
    serviceLoading.value = true;
    try {
      const bizId = props.spaceId
      const query = {
        start: 0,
        limit: 1000, // @todo 确认拉全量列表参数
        operator: userInfo.value.username
      }
      const resp = await getAppList(bizId, query)
      serviceList.value = resp.details
    } catch (e) {
      console.error(e)
    } finally {
        serviceLoading.value = false;
    }
  }

  const change = () => {
    emits('change', localVal.value)
  }

  const validate = () => {
    return formRef.value.validate()
  }

  defineExpose({
    validate
  })

</script>
<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <bk-form-item label="模板套餐名称" property="name" required>
      <bk-input v-model="localVal.name" @change="change" />
    </bk-form-item>
    <bk-form-item label="模板套餐描述" property="memo">
      <bk-input v-model="localVal.memo" type="textarea" :rows="6" :maxlength="256" @change="change" />
    </bk-form-item>
    <bk-form-item label="服务可见范围" property="public" required>
      <bk-radio-group v-model="localVal.public" @change="change">
        <bk-radio :label="true">公开</bk-radio>
        <bk-radio :label="false">指定服务</bk-radio>
      </bk-radio-group>
      <bk-select
        v-if="!localVal.public"
        v-model="localVal.bound_apps"
        class="service-selector"
        multiple
        filterable
        placeholder="请选择服务"
        :loading="serviceLoading"
        @change="change">
        <bk-option v-for="service in serviceList" :key="service.id" :label="service.spec.name" :value="service.id"></bk-option>
      </bk-select>
      <p v-if="deletedApps.length > 0" class="tips">
        提醒：修改可见范围后，服务
        <span v-for="item in deletedApps" :key="item.id">【{{ item.spec.name }}】</span>
        将不再引用此套餐
      </p>
    </bk-form-item>
  </bk-form>
</template>
<style lang="scss" scoped>
  .tips {
    margin: 8px 0;
    line-height: 16px;
    font-size: 12px;
    color: #ff9c01;
  }
</style>
