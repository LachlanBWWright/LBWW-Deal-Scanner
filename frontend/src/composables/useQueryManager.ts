import { computed, onMounted, ref } from "vue";
import {
  supportedQueryTypes,
  type QueryType,
  type QueryItem,
  type QueryFormState,
  type FormField,
  type SearchResultItem,
} from "../components/queryTypes";
import { ResultAsync } from "neverthrow";
import { toError } from "../utils/neverthrowUtils";

const STORAGE_HOST_KEY = "dealscannerApiHost";
const STORAGE_SECRET_KEY = "dealscannerApiSecret";

export function useQueryManager() {
  const apiHost = ref(localStorage.getItem(STORAGE_HOST_KEY) ?? "");
  const apiSecret = ref(localStorage.getItem(STORAGE_SECRET_KEY) ?? "");
  const queries = ref<QueryItem[]>([]);
  const filterType = ref<QueryType | "all">("all");
  const querySearch = ref("");
  const selectedType = ref<QueryType>("ebay");
  const editingQuery = ref<QueryItem | null>(null);
  const searchResults = ref<SearchResultItem[]>([]);
  const lastSyncedAt = ref<string | null>(null);
  const loading = ref(false);
  const saving = ref(false);
  const errorMessage = ref<string | null>(null);
  const successMessage = ref<string | null>(null);

  function initialFormState(): QueryFormState {
    return {
      url: "",
      name: "",
      displayUrl: "",
      maxPrice: 0,
      minPrice: 0,
      minFloat: 0,
      maxFloat: 0,
      requiredPhrases: "",
      excludePhrases: "",
      dmOnly: false,
    };
  }

  const form = ref<QueryFormState>(initialFormState());

  function stringifyForSearch(item: QueryItem) {
    return [
      item.type,
      item.id,
      item.name,
      item.url,
      item.displayUrl,
      item.requiredPhrases,
      item.excludePhrases,
      String(item.maxPrice ?? ""),
      String(item.minPrice ?? ""),
      String(item.minFloat ?? ""),
      String(item.maxFloat ?? ""),
    ]
      .join(" ")
      .toLowerCase();
  }

  const filteredQueries = computed(() => {
    const normalizedSearch = querySearch.value.trim().toLowerCase();
    return queries.value.filter((item) => {
      const typeMatches =
        filterType.value === "all" || item.type === filterType.value;
      if (!typeMatches) return false;

      if (!normalizedSearch) return true;
      return stringifyForSearch(item).includes(normalizedSearch);
    });
  });

  const queryStats = computed(() => {
    return supportedQueryTypes.map((type) => ({
      type,
      count: queries.value.filter((item) => item.type === type).length,
    }));
  });

  const selectedQueryPreview = computed(() => {
    if (editingQuery.value) return editingQuery.value;
    return filteredQueries.value[0] ?? null;
  });

  const formFields = computed<FormField[]>(() => {
    switch (selectedType.value) {
      case "cashConverters":
        return [
          { label: "Query URL", key: "url", type: "url" },
          { label: "Required phrases", key: "requiredPhrases", type: "text" },
          { label: "Exclude phrases", key: "excludePhrases", type: "text" },
        ];
      case "ebay":
      case "gumtree":
        return [
          { label: "Query URL", key: "url", type: "url" },
          { label: "Max price", key: "maxPrice", type: "number" },
        ];
      case "salvos":
        return [
          { label: "Item name", key: "name", type: "text" },
          { label: "Minimum price", key: "minPrice", type: "number" },
          { label: "Maximum price", key: "maxPrice", type: "number" },
        ];
      case "csMarket":
        return [
          { label: "Search URL", key: "url", type: "url" },
          { label: "Display URL", key: "displayUrl", type: "text" },
          { label: "Max price", key: "maxPrice", type: "number" },
          { label: "Max float", key: "maxFloat", type: "number" },
        ];
      case "steamMarket":
        return [
          { label: "Query name", key: "name", type: "text" },
          { label: "Display URL", key: "displayUrl", type: "text" },
          { label: "Max price", key: "maxPrice", type: "number" },
        ];
      case "csTradeBot":
        return [
          { label: "Item name", key: "name", type: "text" },
          { label: "Min float", key: "minFloat", type: "number" },
          { label: "Max float", key: "maxFloat", type: "number" },
          { label: "Max price", key: "maxPrice", type: "number" },
        ];
      default:
        return [];
    }
  });

  function getApiUrl(path: string) {
    const cleanedHost = apiHost.value?.replace(/\/+$/, "") ?? "";
    return cleanedHost ? `${cleanedHost}${path}` : path;
  }

  function resetForm() {
    editingQuery.value = null;
    form.value = initialFormState();
  }

  function selectType(type: QueryType) {
    selectedType.value = type;
    resetForm();
  }

  function populateForm(item: QueryItem) {
    selectedType.value = item.type;
    editingQuery.value = item;
    form.value = {
      url: item.url || "",
      name: item.name || "",
      displayUrl: item.displayUrl || "",
      maxPrice: item.maxPrice ?? 0,
      minPrice: item.minPrice ?? 0,
      minFloat: item.minFloat ?? 0,
      maxFloat: item.maxFloat ?? 0,
      requiredPhrases: item.requiredPhrases || "",
      excludePhrases: item.excludePhrases || "",
      dmOnly: item.dmOnly,
    };
  }

  function buildPayload() {
    const payload: Record<string, unknown> = {
      dmOnly: form.value.dmOnly,
    };

    switch (selectedType.value) {
      case "cashConverters":
        payload.url = form.value.url;
        payload.requiredPhrases = form.value.requiredPhrases;
        payload.excludePhrases = form.value.excludePhrases;
        break;
      case "ebay":
      case "gumtree":
        payload.url = form.value.url;
        payload.maxPrice = form.value.maxPrice;
        break;
      case "salvos":
        payload.name = form.value.name;
        payload.minPrice = form.value.minPrice;
        payload.maxPrice = form.value.maxPrice;
        break;
      case "csMarket":
        payload.url = form.value.url;
        payload.displayUrl = form.value.displayUrl || form.value.url;
        payload.maxPrice = form.value.maxPrice;
        payload.maxFloat = form.value.maxFloat;
        break;
      case "steamMarket":
        payload.name = form.value.name;
        payload.displayUrl = form.value.displayUrl || form.value.name;
        payload.maxPrice = form.value.maxPrice;
        break;
      case "csTradeBot":
        payload.name = form.value.name;
        payload.minFloat = form.value.minFloat;
        payload.maxFloat = form.value.maxFloat;
        payload.maxPrice = form.value.maxPrice;
        break;
    }

    return payload;
  }

  function buildFetchHeaders() {
    return {
      "Content-Type": "application/json",
      ...(apiSecret.value ? { "x-api-secret": apiSecret.value } : {}),
    } as HeadersInit;
  }

  async function refreshQueries() {
    loading.value = true;
    errorMessage.value = null;
    successMessage.value = null;

    const responseResult = await ResultAsync.fromThrowable(
      () =>
        fetch(getApiUrl("/api/queries"), {
          method: "GET",
          headers: buildFetchHeaders(),
        }),
      (error) => toError(error, "Failed to load saved queries"),
    )();

    if (responseResult.isErr()) {
      errorMessage.value = responseResult.error.message;
      loading.value = false;
      return;
    }

    const response = responseResult.value;
    if (!response.ok) {
      errorMessage.value = `Failed to load saved queries (${response.status})`;
      loading.value = false;
      return;
    }

    const dataResult = await ResultAsync.fromThrowable(
      () => response.json() as Promise<{ queries: QueryItem[] }>,
      (error) => toError(error, "Failed to parse saved queries"),
    )();
    if (dataResult.isErr()) {
      errorMessage.value = dataResult.error.message;
      loading.value = false;
      return;
    }

    queries.value = dataResult.value.queries ?? [];
    lastSyncedAt.value = new Date().toISOString();

    await refreshSearchResults();

    loading.value = false;
  }

  async function refreshSearchResults() {
    const responseResult = await ResultAsync.fromThrowable(
      () =>
        fetch(getApiUrl("/api/search-results"), {
          method: "GET",
          headers: buildFetchHeaders(),
        }),
      (error) => toError(error, "Failed to load recent search results"),
    )();

    if (responseResult.isErr()) {
      errorMessage.value = responseResult.error.message;
      return;
    }

    const response = responseResult.value;
    if (!response.ok) {
      errorMessage.value = `Failed to load recent search results (${response.status})`;
      return;
    }

    const dataResult = await ResultAsync.fromThrowable(
      () => response.json() as Promise<{ results: SearchResultItem[] }>,
      (error) => toError(error, "Failed to parse recent search results"),
    )();
    if (dataResult.isErr()) {
      errorMessage.value = dataResult.error.message;
      return;
    }

    searchResults.value = dataResult.value.results ?? [];
  }

  async function saveQuery() {
    saving.value = true;
    errorMessage.value = null;
    successMessage.value = null;

    const body = {
      type: selectedType.value,
      payload: buildPayload(),
    };

    const responseResult = await ResultAsync.fromThrowable(
      () =>
        fetch(getApiUrl("/api/queries"), {
          method: editingQuery.value ? "PUT" : "POST",
          headers: buildFetchHeaders(),
          body: JSON.stringify(
            editingQuery.value ? { ...body, id: editingQuery.value.id } : body,
          ),
        }),
      (error) => toError(error, "Failed to save query"),
    )();

    if (responseResult.isErr()) {
      errorMessage.value = responseResult.error.message;
      saving.value = false;
      return;
    }

    const response = responseResult.value;
    if (!response.ok) {
      errorMessage.value = `Failed to save query (${response.status})`;
      saving.value = false;
      return;
    }

    successMessage.value = editingQuery.value
      ? "Query updated successfully."
      : "Query created successfully.";

    resetForm();
    const refreshResult = await ResultAsync.fromThrowable(
      () => refreshQueries(),
      (error) => toError(error, "Failed to save query"),
    )();
    if (refreshResult.isErr()) {
      errorMessage.value = refreshResult.error.message;
      saving.value = false;
      return;
    }

    saving.value = false;
  }

  async function deleteQuery(item: QueryItem) {
    if (
      !window.confirm(
        `Delete ${item.type} query ${item.id}? This cannot be undone.`,
      )
    ) {
      return;
    }

    saving.value = true;
    errorMessage.value = null;
    successMessage.value = null;

    const responseResult = await ResultAsync.fromThrowable(
      () =>
        fetch(getApiUrl("/api/queries"), {
          method: "DELETE",
          headers: buildFetchHeaders(),
          body: JSON.stringify({ type: item.type, id: item.id }),
        }),
      (error) => toError(error, "Failed to delete query"),
    )();

    if (responseResult.isErr()) {
      errorMessage.value = responseResult.error.message;
      saving.value = false;
      return;
    }

    const response = responseResult.value;
    if (!response.ok) {
      errorMessage.value = `Failed to delete query (${response.status})`;
      saving.value = false;
      return;
    }

    successMessage.value = "Query deleted successfully.";
    if (
      editingQuery.value?.id === item.id &&
      editingQuery.value.type === item.type
    ) {
      resetForm();
    }

    const refreshResult = await ResultAsync.fromThrowable(
      () => refreshQueries(),
      (error) => toError(error, "Failed to delete query"),
    )();
    if (refreshResult.isErr()) {
      errorMessage.value = refreshResult.error.message;
      saving.value = false;
      return;
    }

    saving.value = false;
  }

  function setApiHost() {
    localStorage.setItem(STORAGE_HOST_KEY, apiHost.value);
    localStorage.setItem(STORAGE_SECRET_KEY, apiSecret.value);
    successMessage.value = apiHost.value
      ? `Using API host ${apiHost.value}`
      : `Using current host`;
    if (apiSecret.value) {
      successMessage.value += " with API secret.";
    }
    errorMessage.value = null;
    void refreshQueries();
  }

  function clearApiHost() {
    apiHost.value = "";
    apiSecret.value = "";
    localStorage.removeItem(STORAGE_HOST_KEY);
    localStorage.removeItem(STORAGE_SECRET_KEY);
    successMessage.value = "Using current host";
    errorMessage.value = null;
    void refreshQueries();
  }

  function startEditing(item: QueryItem) {
    populateForm(item);
    successMessage.value = null;
    errorMessage.value = null;
  }

  onMounted(() => {
    void refreshQueries();
  });

  return {
    supportedQueryTypes,
    apiHost,
    apiSecret,
    querySearch,
    queryStats,
    selectedQueryPreview,
    lastSyncedAt,
    filterType,
    selectedType,
    editingQuery,
    loading,
    saving,
    errorMessage,
    successMessage,
    form,
    filteredQueries,
    searchResults,
    formFields,
    setApiHost,
    clearApiHost,
    selectType,
    saveQuery,
    resetForm,
    refreshQueries,
    refreshSearchResults,
    startEditing,
    deleteQuery,
  };
}
