export interface paths {
    "/api/status": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Read runtime status */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            api: {
                                startedAt: string;
                                port: number;
                            };
                            bot: {
                                /** @enum {string} */
                                status: "starting" | "ready" | "disabled" | "error";
                                connected: boolean;
                                lastReadyAt: null | string;
                                lastError: null | string;
                            };
                            scanner: {
                                /** @enum {string} */
                                status: "idle" | "running" | "error";
                                loopRunning: boolean;
                                lastRun: null | {
                                    /** @enum {string} */
                                    mode: "loop" | "manual";
                                    startedAt: string;
                                    finishedAt: string | null;
                                    durationMs: number | null;
                                    error: string | null;
                                };
                            };
                            commands: string[];
                        };
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/commands": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List registered Discord commands */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            commands: string[];
                        };
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/scans/run": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Run a manual scan pass */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody: {
                content: {
                    "application/json": {
                        reason?: string;
                    };
                };
            };
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            accepted: boolean;
                            startedAt: string;
                            finishedAt: string;
                            durationMs: number;
                            /** @enum {string} */
                            mode: "manual";
                            error: null | string;
                        };
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/search-results": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List recent scanner search results */
        get: {
            parameters: {
                query?: {
                    type?: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                    queryId?: string;
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            results: {
                                source: string;
                                title: string;
                                url: string;
                                price: null | number;
                                imageUrl: null | string;
                                queryType: null | string;
                                queryId: null | string;
                                foundAt: string;
                            }[];
                        };
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/queries": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List saved queries */
        get: {
            parameters: {
                query?: {
                    type?: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                };
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            queries: {
                                /** @enum {string} */
                                type: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                                id: string;
                                dmOnly: boolean;
                                url?: string;
                                name?: string;
                                displayUrl?: string;
                                maxPrice?: number;
                                minPrice?: number;
                                maxFloat?: number;
                                requiredPhrases?: string;
                                excludePhrases?: string;
                            }[];
                        };
                    };
                };
            };
        };
        /** Update a saved query */
        put: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody: {
                content: {
                    "application/json": {
                        /** @enum {string} */
                        type: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                        id: string;
                        payload: {
                            [key: string]: unknown;
                        };
                    };
                };
            };
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            success: boolean;
                            query: {
                                /** @enum {string} */
                                type: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                                id: string;
                                dmOnly: boolean;
                                url?: string;
                                name?: string;
                                displayUrl?: string;
                                maxPrice?: number;
                                minPrice?: number;
                                maxFloat?: number;
                                requiredPhrases?: string;
                                excludePhrases?: string;
                            };
                        };
                    };
                };
            };
        };
        /** Create a saved query */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody: {
                content: {
                    "application/json": {
                        /** @enum {string} */
                        type: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                        payload: {
                            [key: string]: unknown;
                        };
                    };
                };
            };
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            success: boolean;
                            query: {
                                /** @enum {string} */
                                type: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                                id: string;
                                dmOnly: boolean;
                                url?: string;
                                name?: string;
                                displayUrl?: string;
                                maxPrice?: number;
                                minPrice?: number;
                                maxFloat?: number;
                                requiredPhrases?: string;
                                excludePhrases?: string;
                            };
                        };
                    };
                };
            };
        };
        /** Delete a saved query */
        delete: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody: {
                content: {
                    "application/json": {
                        /** @enum {string} */
                        type: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                        id: string;
                    };
                };
            };
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            success: boolean;
                        };
                    };
                };
            };
        };
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/testing/capabilities": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get testing API capabilities */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            testingEnabled: boolean;
                            notificationProviders: string[];
                            discordConnected: boolean;
                            availableQueryTypes: string[];
                            availableScannerTypes: string[];
                        };
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/testing/notifications": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List recent test notification history */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            results: {
                                id: string;
                                startedAt: string;
                                finishedAt: string;
                                durationMs: number;
                                outcomes: {
                                    provider: string;
                                    /** @enum {string} */
                                    status: "sent" | "skipped" | "disabled" | "failed";
                                    reason?: null | string;
                                }[];
                                error: null | string;
                            }[];
                        };
                    };
                };
            };
        };
        put?: never;
        /** Send a test notification through NotificationService */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody: {
                content: {
                    "application/json": {
                        /** @enum {string} */
                        kind: "deal" | "error";
                        source?: string;
                        title?: string;
                        url?: string;
                        price?: number;
                        imageUrl?: string;
                        query?: {
                            /** @enum {string} */
                            type: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                            id: string;
                            dmOnly?: boolean;
                        };
                        tags?: string[];
                        message?: string;
                        /** @enum {string} */
                        deliveryMode?: "normal" | "guildChannelOnly" | "subscribedDMs" | "specificUserDM";
                        targetDiscordUserId?: string;
                    };
                };
            };
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            id: string;
                            startedAt: string;
                            finishedAt: string;
                            durationMs: number;
                            outcomes: {
                                provider: string;
                                /** @enum {string} */
                                status: "sent" | "skipped" | "disabled" | "failed";
                                reason?: null | string;
                            }[];
                            error: null | string;
                        };
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/testing/scans/run": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Run a temporary scan without persisting a user query */
        post: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody: {
                content: {
                    "application/json": {
                        /** @enum {string} */
                        type: "cashConverters" | "ebay" | "gumtree" | "salvos" | "csMarket" | "steamMarket" | "csTradeBot";
                        payload: {
                            [key: string]: unknown;
                        };
                        notify?: boolean;
                    };
                };
            };
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            id: string;
                            startedAt: string;
                            finishedAt: string;
                            durationMs: number;
                            notifications: {
                                source: string;
                                title: string;
                                url: string;
                                price: null | number;
                                imageUrl: null | string;
                            }[];
                            errors: string[];
                            notificationsPublished: boolean;
                        };
                    };
                };
            };
        };
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/api/testing/runs": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List recent temporary scan run history */
        get: {
            parameters: {
                query?: never;
                header?: never;
                path?: never;
                cookie?: never;
            };
            requestBody?: never;
            responses: {
                /** @description Default Response */
                200: {
                    headers: {
                        [name: string]: unknown;
                    };
                    content: {
                        "application/json": {
                            results: {
                                id: string;
                                startedAt: string;
                                finishedAt: string;
                                durationMs: number;
                                notifications: {
                                    source: string;
                                    title: string;
                                    url: string;
                                    price: null | number;
                                    imageUrl: null | string;
                                }[];
                                errors: string[];
                                notificationsPublished: boolean;
                            }[];
                        };
                    };
                };
            };
        };
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: never;
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export type operations = Record<string, never>;

