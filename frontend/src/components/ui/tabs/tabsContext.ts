import type { InjectionKey, Ref } from "vue";

export type TabsContext = {
  readonly value: Ref<string>;
  readonly setValue: (nextValue: string) => void;
};

export const tabsKey: InjectionKey<TabsContext> = Symbol("Tabs");
