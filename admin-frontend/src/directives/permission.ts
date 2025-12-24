import type {DirectiveBinding} from 'vue';
import {usePermission} from '@/hooks/usePermission';

export default {
  mounted(el: HTMLElement, binding: DirectiveBinding<string | string[]>) {
    const {hasPermission} = usePermission();
    const value = binding.value;
    const required = Array.isArray(value) ? value : [value].filter(Boolean);
    if (!required.length) return;
    const pass = required.some((code) => hasPermission(code));
    if (!pass) {
      el.parentNode?.removeChild(el);
    }
  }
};

