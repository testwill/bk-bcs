<script setup lang="ts">
  import { ref, computed, watch, onMounted } from 'vue'
  import SHA256 from 'crypto-js/sha256'
  import WordArray from 'crypto-js/lib-typedarrays'
  import { TextFill, Done } from 'bkui-vue/lib/icon'
  import BkMessage from 'bkui-vue/lib/message'
  import { IConfigEditParams, IFileConfigContentSummary } from '../../../../../../../../types/config'
  import { updateConfigContent, getConfigContent } from '../../../../../../../api/config'
  import { stringLengthInBytes } from '../../../../../../../utils/index'
  import { transFileToObject, fileDownload } from '../../../../../../../utils/file'
  import { CONFIG_FILE_TYPE } from '../../../../../../../constants/config'
  import ConfigContentEditor from '../../components/config-content-editor.vue'

  const PRIVILEGE_GROUPS = ['属主（own）', '属组（group）', '其他人（other）']
  const PRIVILEGE_VALUE_MAP = {
    0: [],
    1: [1],
    2: [2],
    3: [1, 2],
    4: [4],
    5: [1,4],
    6: [2, 4],
    7: [1, 2, 4]
  }

  const props = withDefaults(defineProps<{
    config: IConfigEditParams;
    editable: boolean;
    content: string|IFileConfigContentSummary;
    bkBizId: string;
    appId: number;
    fileUploading?: boolean;
  }>(), {
    editable: true
  })

  const emits = defineEmits(['change', 'update:fileUploading'])

  const localVal = ref({ ...props.config })
  const privilegeInputVal = ref('')
  const showPrivilegeErrorTips = ref(false)
  const stringContent = ref('')
  const fileContent = ref<IFileConfigContentSummary|File>()
  const isFileChanged = ref(false) // 标识文件是否被修改，编辑配置项时若文件未修改，不重新上传文件
  const uploadPending = ref(false)
  const formRef = ref()
  const rules = {
    name: [
      {
        validator: (value: string) => value.length <= 64,
        message: '最大长度64个字符'
      },
      {
        validator: (value: string) => {
          return /^[a-zA-Z0-9_\-\.]+$/.test(value)
        },
        message: '请使用英文、数字、下划线、中划线或点'
      }
    ],
    privilege: [{
      required: true,
      validator: () => {
        const type = typeof privilegeInputVal.value
        return type === 'number' || (type === 'string' && privilegeInputVal.value.length > 0)
      },
      message: '文件权限 不能为空',
      trigger: 'change'
    }],
    path: [
      {
        validator: (value: string) => value.length <= 256,
        message: '最大长度256个字符'
      }
    ],
  }

  // 传入到bk-upload组件的文件对象
  const fileList = computed(() => {
    return fileContent.value ? [transFileToObject(<File>fileContent.value)] : []
  })

  // 将权限数字拆分成三个分组配置
  const privilegeGroupsValue = computed(() => {
    let data: { [index: string]: number[] } = { '0': [], '1': [], '2': [] }
    if (typeof localVal.value.privilege === 'string' && localVal.value.privilege.length > 0) {
      const valArr = localVal.value.privilege.split('').map(i => parseInt(i))
      valArr.forEach((item, index) => {
        data[index as keyof typeof data] = PRIVILEGE_VALUE_MAP[item as keyof typeof PRIVILEGE_VALUE_MAP]
      })
    }
    return data
  })

  watch(() => props.config.privilege, (val) => {
    privilegeInputVal.value = <string>val
  }, { immediate: true })

  watch(() => props.content, () => {
    if (props.config.file_type === 'binary') {
      fileContent.value = <IFileConfigContentSummary>props.content
    } else {
      stringContent.value = props.content as string
    }
  }, { immediate: true })

  // 权限输入框失焦后，校验输入是否合法，如不合法回退到上次输入
  const handlePrivilegeInputBlur = () => {
    const val = String(privilegeInputVal.value)
    if (/^[0-7]{3}$/.test(val)) {
      localVal.value.privilege = val
      showPrivilegeErrorTips.value = false
      change()
    } else {
      privilegeInputVal.value = String(localVal.value.privilege)
      showPrivilegeErrorTips.value = true
    }
  }

  // 选择文件权限
  const handleSelectPrivilege = (index: number, val: number[]) => {
    const groupsValue = { ...privilegeGroupsValue.value }
    groupsValue[index] = val
    const digits = []
    for(let i = 0; i < 3; i++) {
      let sum = 0
      if (groupsValue[i].length > 0) {
        sum = groupsValue[i].reduce((acc, crt) => {
          return acc + crt
        }, 0)
      }
      digits.push(sum)
    }
    const newVal = digits.join('')
    privilegeInputVal.value = newVal
    localVal.value.privilege = newVal
    showPrivilegeErrorTips.value = false
    change()
  }

  const handleStringContentChange = (val: string) => {
    stringContent.value = val
    change()
  }

  // 选择文件后上传
  const handleFileUpload = (option: { file: File }) => {
    isFileChanged.value = true
    return new Promise(resolve => {
      uploadPending.value = true
      emits('update:fileUploading', true)
      fileContent.value = option.file
      uploadContent().then(res => {
        uploadPending.value = false
        emits('update:fileUploading', false)
        change()
        resolve(res)
      })
    })
  }

  // 上传配置内容
  const uploadContent =  async () => {
    const signature = await getSignature()
    return updateConfigContent(props.bkBizId, props.appId, <File>fileContent.value, <string>signature)
  }

  // 生成文件或文本的sha256
  const getSignature = async () => {
    if (localVal.value.file_type === 'binary') {
      if (isFileChanged.value) {
        return new Promise(resolve => {
          const reader = new FileReader()
          // @ts-ignore
          reader.readAsArrayBuffer(fileContent.value)
          reader.onload = () => {
            const wordArray = WordArray.create(reader.result);
            resolve(SHA256(wordArray).toString())
          }
        })
      }
      return (fileContent.value as IFileConfigContentSummary).signature
    }
    return SHA256(stringContent.value).toString()
  }

  // 下载已上传文件
  const handleDownloadFile = async () => {
    const { signature, name } = <IFileConfigContentSummary>fileContent.value
    const res = await getConfigContent(props.bkBizId, props.appId, signature)
    fileDownload(res, `${name}.bin`)
  }

  const validate = async() => {
    await formRef.value.validate()
    if (localVal.value.file_type === 'binary'){
      if (fileList.value.length === 0) {
        BkMessage({ theme: 'error', message: '请上传文件' })
        return false
      }
    } else if (localVal.value.file_type === 'text') {
      if (stringLengthInBytes(stringContent.value) > 1024 * 1024 * 50 ) {
        BkMessage({ theme: 'error', message: '配置内容不能超过50M' })
        return false
      }
    }
    return true
  }

  const change = () => {
    const content = localVal.value.file_type === 'binary' ? fileContent.value : stringContent.value
    emits('change', localVal.value, content)
  }

  defineExpose({
    getSignature,
    validate
  })

</script>
<template>
  <bk-form ref="formRef" form-type="vertical" :model="localVal" :rules="rules">
    <bk-form-item label="配置项名称" property="name" :required="true">
      <bk-input
        v-model="localVal.name"
        placeholder="请输入1~64个字符，只允许英文、数字、下划线、中划线或点"
        :disabled="!editable"
        @change="change" />
    </bk-form-item>
    <bk-form-item label="配置项路径" property="path" :required="true">
      <bk-input
        v-model="localVal.path"
        placeholder="请输入绝对路径，下载路径为前缀+配置路径"
        :disabled="!editable"
        @change="change" />
    </bk-form-item>
    <bk-form-item label="配置项描述" property="memo">
      <bk-input v-model="localVal.memo" type="textarea" :disabled="!editable" @change="change"></bk-input>
    </bk-form-item>
    <bk-form-item label="配置项格式">
    <bk-radio-group v-model="localVal.file_type" :required="true" @change="change">
        <bk-radio v-for="typeItem in CONFIG_FILE_TYPE" :key="typeItem.id" :label="typeItem.id" :disabled="!editable">{{ typeItem.name }}</bk-radio>
    </bk-radio-group>
    </bk-form-item>
    <div class="user-settings">
      <bk-form-item label="文件权限" property="privilege" required>
        <div class="perm-input">
          <bk-popover
            theme="light"
            trigger="manual"
            placement="top"
            :is-show="showPrivilegeErrorTips">
            <bk-input
              v-model="privilegeInputVal"
              type="number"
              placeholder="请输入三位权限数字"
              :disabled="!editable"
              @blur="handlePrivilegeInputBlur" />
            <template #content>
              <div>只能输入三位 0~7 数字</div>
              <div class="privilege-tips-btn-area">
                <bk-button text theme="primary" @click="showPrivilegeErrorTips = false">我知道了</bk-button>
              </div>
            </template>
          </bk-popover>
          <bk-popover
            ext-cls="privilege-select-popover"
            theme="light"
            trigger="click"
            placement="bottom"
            :disabled="!editable">
            <div :class="['perm-panel-trigger', { disabled: !editable }]">
              <i class="bk-bscp-icon icon-configuration-line"></i>
            </div>
            <template #content>
              <div class="privilege-select-panel">
                <div v-for="(item, index) in PRIVILEGE_GROUPS" class="group-item" :key="index" :label="item">
                  <div class="header">{{ item }}</div>
                  <div class="checkbox-area">
                    <bk-checkbox-group
                      class="group-checkboxs"
                      :model-value="privilegeGroupsValue[index]"
                      @change="handleSelectPrivilege(index, $event)">
                      <bk-checkbox size="small" :label="4">读</bk-checkbox>
                      <bk-checkbox size="small" :label="2">写</bk-checkbox>
                      <bk-checkbox size="small" :label="1">执行</bk-checkbox>
                    </bk-checkbox-group>
                  </div>
                </div>
              </div>
            </template>
          </bk-popover>
        </div>
      </bk-form-item>
      <bk-form-item label="用户" property="user" :required="true">
        <bk-input v-model="localVal.user" :disabled="!editable" @change="change"></bk-input>
      </bk-form-item>
      <bk-form-item label="用户组" property="user_group" :required="true">
        <bk-input v-model="localVal.user_group" :disabled="!editable" @change="change"></bk-input>
      </bk-form-item>
    </div>
    <bk-form-item v-if="localVal.file_type === 'binary'" label="配置内容" :required="true">
      <bk-upload
        class="config-uploader"
        url=""
        theme="button"
        tip="支持扩展名：.bin，文件大小100M以内"
        :size="100"
        :disabled="!editable"
        :multiple="false"
        :files="fileList"
        :custom-request="handleFileUpload">
        <template #file="{ file }">
          <div class="file-wrapper">
            <Done class="done-icon"/>
            <TextFill class="file-icon" />
            <div v-bk-ellipsis class="name" @click="handleDownloadFile">{{ file.name }}</div>
            ({{ file.size }})
          </div>
        </template>
      </bk-upload>
    </bk-form-item>
    <bk-form-item v-else label="配置内容" :required="true">
      <ConfigContentEditor
        :content="stringContent"
        :editable="editable"
        @change="handleStringContentChange" />
    </bk-form-item>
  </bk-form>
</template>
<style lang="scss" scoped>
  .user-settings {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
  }
  .perm-input {
    display: flex;
    align-items: center;
    width: 172px;
    :deep(.bk-input) {
      width: 140px;
      border-right: none;
      border-top-right-radius: 0;
      border-bottom-right-radius: 0;
      .bk-input--number-control {
        display: none;
      }
    }
    .perm-panel-trigger {
      width: 32px;
      height: 32px;
      text-align: center;
      background: #e1ecff;
      color: #3a84ff;
      border: 1px solid #3a84ff;
      cursor: pointer;
      &.disabled {
        color: #dcdee5;
        border-color: #dcdee5;
        cursor: not-allowed;
      }
    }
  }
  .privilege-tips-btn-area {
    margin-top: 8px;
    text-align: right;
  }
  .privilege-select-panel {
    display: flex;
    align-items: top;
    border: 1px solid #dcdee5;
    .group-item {
      .header {
        padding: 0 16px;
        height: 42px;
        line-height: 42px;
        color: #313238;
        font-size: 12px;
        background: #fafbfd;
        border-bottom: 1px solid #dcdee5;
      }
      &:not(:last-of-type) {
        .header, .checkbox-area {
          border-right: 1px solid #dcdee5;
        }
      }
    }
    .checkbox-area {
      padding: 10px 16px 12px;
      background: #ffffff;
      &:not(:last-child) {
        border-right: 1px solid #dcdee5;
      }
    }
    .group-checkboxs {
      font-size: 12px;
      .bk-checkbox~.bk-checkbox {
        margin-left: 16px;
      }
      :deep(.bk-checkbox-label) {
        font-size: 12px;
      }
    }
  }
  .config-uploader {
    :deep(.bk-upload-list__item) {
      padding: 0;
      border: none;
    }
    :deep(.bk-upload-list--disabled .bk-upload-list__item) {
      pointer-events: inherit;
    }
    .file-wrapper {
      display: flex;
      align-items: center;
      color: #979ba5;
      font-size: 12px;
      .done-icon {
        font-size: 20px;
        color: #2dcb56;
      }
      .file-icon {
        margin: 0 6px 0 0;
      }
      .name {
        max-width: 360px;
        margin-right: 4px;
        color: #63656e;
        white-space: nowrap;
        text-overflow: ellipsis;
        overflow: hidden;
        cursor: pointer;
        &:hover {
          color: #3a84ff;
          text-decoration: underline;
        }
      }
    }
  }
</style>
<style lang="scss">
  .privilege-select-popover.bk-popover {
    padding: 0;
  }
</style>
