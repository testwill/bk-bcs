<script lang="ts" setup>
  import { Ellipsis } from 'bkui-vue/lib/icon'
  import { IPackageMenuItem } from '../../../../../../types/template'

  const props = defineProps<{
    currentPkg: number|string;
    pkg: IPackageMenuItem;
    hideActions?: boolean;
  }>()

  const emits = defineEmits(['edit', 'clone', 'delete', 'select'])

</script>
<template>
  <div :class="['package-item', { active: props.pkg.id === props.currentPkg }]" @click="emits('select', props.pkg.id)">
    <div class="pkg-wrapper">
      <div class="mark-icon">
        <slot name="icon">
          <i class="bk-bscp-icon icon-folder"></i>
        </slot>
      </div>
      <div class="text">
        <span class="name">{{ props.pkg.name }}</span>
        <span class="num">{{ props.pkg.count }}</span>
      </div>
      <bk-popover
        v-if="typeof props.pkg.id === 'number'"
        theme="light template-package-actions-popover"
        placement="bottom-end"
        :arrow="false">
        <Ellipsis class="action-more-icon" @click.stop />
        <template #content>
          <div class="package-actions">
            <div class="action-item" @click="emits('edit', props.pkg.id, 'edit')">编辑</div>
            <div class="action-item" @click="emits('clone', props.pkg.id, 'clone')">克隆</div>
            <div class="action-item" @click="emits('delete', props.pkg.id, 'delete')">删除</div>
          </div>
        </template>
      </bk-popover>
    </div>
  </div>
</template>
<style lang="scss" scoped>
  .package-item {
    padding: 8px 16px;
    cursor: pointer;
    &.active {
      background: #e1ecff;
      .mark-icon {
        color: #3a84ff;
      }
      .text {
        .name {
          color: #3a84ff;
        }
        .num {
          background: #a3c5fd;
          color: #ffffff;
        }
      }
    }
    .pkg-wrapper {
      display: flex;
      align-items: center;
    }
    .mark-icon {
      display: flex;
      align-items: center;
      width: 12px;
      height: 12px;
      color: #c4c6cc;
      font-size: 12px;
      .icon-folder {
        transform-origin: 0 50%;
        transform: scale(0.8);
      }
      .icon-empty {
        transform-origin: 0 50%;
        transform: scale(0.7);
      }
    }
    .text {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0 4px;
      width: calc(100% - 30px);
      .name {
        font-size: 12px;
        color: #63656e;
      }
      .num {
        padding: 0 8px;
        color: #979ba5;
        height: 16px;
        line-height: 16px;
        font-size: 12px;
        background: #f0f1f5;
        border-radius: 2px;
      }
    }
    .action-more-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      transform: rotate(90deg);
      width: 16px;
      height: 16px;
      color: #979ba5;
      border-radius: 50%;
      cursor: pointer;
      &:hover {
        background: rgba(99, 101, 110, 0.1);
        color: #3a84ff;
      }
    }
  }
</style>
<style lang="scss">
  .template-package-actions-popover.bk-popover.bk-pop2-content {
    padding: 4px 0;
    border: 1px solid #dcdee5;
    box-shadow: 0 2px 6px 0 #0000001a;
    .package-actions {
      .action-item {
        padding: 0 12px;
        min-width: 58px;
        height: 32px;
        line-height: 32px;
        color: #63656e;
        font-size: 12px;
        cursor: pointer;
        &:hover {
          background: #f5f7fa;
        }
      }
    }
  }
</style>
