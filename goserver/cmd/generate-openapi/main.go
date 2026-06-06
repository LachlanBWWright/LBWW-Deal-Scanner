package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const openApiJSON = `{
  "openapi": "3.1.0",
  "info": {
    "title": "DealScanner Control API",
    "description": "Fastify control API for the DealScanner bot and scanner runtime.",
    "version": "1.0.0"
  },
  "components": {
    "schemas": {}
  },
  "paths": {
    "/api/status": {
      "get": {
        "summary": "Read runtime status",
        "tags": [
          "control"
        ],
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "api": {
                      "type": "object",
                      "additionalProperties": false,
                      "properties": {
                        "startedAt": {
                          "type": "string"
                        },
                        "port": {
                          "type": "number"
                        }
                      },
                      "required": [
                        "startedAt",
                        "port"
                      ]
                    },
                    "bot": {
                      "type": "object",
                      "additionalProperties": false,
                      "properties": {
                        "status": {
                          "type": "string",
                          "enum": [
                            "starting",
                            "ready",
                            "disabled",
                            "error"
                          ]
                        },
                        "connected": {
                          "type": "boolean"
                        },
                        "lastReadyAt": {
                          "type": [
                            "null",
                            "string"
                          ]
                        },
                        "lastError": {
                          "type": [
                            "null",
                            "string"
                          ]
                        }
                      },
                      "required": [
                        "status",
                        "connected",
                        "lastReadyAt",
                        "lastError"
                      ]
                    },
                    "scanner": {
                      "type": "object",
                      "additionalProperties": false,
                      "properties": {
                        "status": {
                          "type": "string",
                          "enum": [
                            "idle",
                            "running",
                            "error"
                          ]
                        },
                        "loopRunning": {
                          "type": "boolean"
                        },
                        "lastRun": {
                          "anyOf": [
                            {
                              "type": "null"
                            },
                            {
                              "type": "object",
                              "additionalProperties": false,
                              "properties": {
                                "mode": {
                                  "type": "string",
                                  "enum": [
                                    "loop",
                                    "manual"
                                  ]
                                },
                                "startedAt": {
                                  "type": "string"
                                },
                                "finishedAt": {
                                  "type": [
                                    "string",
                                    "null"
                                  ]
                                },
                                "durationMs": {
                                  "type": [
                                    "number",
                                    "null"
                                  ]
                                },
                                "error": {
                                  "type": [
                                    "string",
                                    "null"
                                  ]
                                }
                              },
                              "required": [
                                "mode",
                                "startedAt",
                                "finishedAt",
                                "durationMs",
                                "error"
                              ]
                            }
                          ]
                        }
                      },
                      "required": [
                        "status",
                        "loopRunning",
                        "lastRun"
                      ]
                    },
                    "commands": {
                      "type": "array",
                      "items": {
                        "type": "string"
                      }
                    }
                  },
                  "required": [
                    "api",
                    "bot",
                    "scanner",
                    "commands"
                  ]
                }
              }
            }
          }
        }
      }
    },
    "/api/commands": {
      "get": {
        "summary": "List registered Discord commands",
        "tags": [
          "control"
        ],
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "commands": {
                      "type": "array",
                      "items": {
                        "type": "string"
                      }
                    }
                  },
                  "required": [
                    "commands"
                  ]
                }
              }
            }
          }
        }
      }
    },
    "/api/scans/run": {
      "post": {
        "summary": "Run a manual scan pass",
        "tags": [
          "control"
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "additionalProperties": false,
                "properties": {
                  "reason": {
                    "type": "string"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "accepted": {
                      "type": "boolean"
                    },
                    "startedAt": {
                      "type": "string"
                    },
                    "finishedAt": {
                      "type": "string"
                    },
                    "durationMs": {
                      "type": "number"
                    },
                    "mode": {
                      "type": "string",
                      "enum": [
                        "manual"
                      ]
                    },
                    "error": {
                      "type": [
                        "null",
                        "string"
                      ]
                    }
                  },
                  "required": [
                    "accepted",
                    "startedAt",
                    "finishedAt",
                    "durationMs",
                    "mode",
                    "error"
                  ]
                }
              }
            }
          }
        }
      }
    },
    "/api/search-results": {
      "get": {
        "summary": "List recent scanner search results",
        "tags": [
          "control"
        ],
        "parameters": [
          {
            "schema": {
              "type": "string",
              "enum": [
                "cashConverters",
                "ebay",
                "gumtree",
                "salvos",
                "csMarket",
                "steamMarket",
                "csTradeBot"
              ]
            },
            "in": "query",
            "name": "type",
            "required": false
          },
          {
            "schema": {
              "type": "string"
            },
            "in": "query",
            "name": "queryId",
            "required": false
          }
        ],
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "results": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "additionalProperties": false,
                        "properties": {
                          "source": {
                            "type": "string"
                          },
                          "title": {
                            "type": "string"
                          },
                          "url": {
                            "type": "string"
                          },
                          "price": {
                            "type": [
                              "number",
                              "null"
                            ]
                          },
                          "imageUrl": {
                            "type": [
                              "string",
                              "null"
                            ]
                          },
                          "queryType": {
                            "type": [
                              "string",
                              "null"
                            ]
                          },
                          "queryId": {
                            "type": [
                              "string",
                              "null"
                            ]
                          },
                          "foundAt": {
                            "type": "string"
                          }
                        },
                        "required": [
                          "source",
                          "title",
                          "url",
                          "price",
                          "imageUrl",
                          "queryType",
                          "queryId",
                          "foundAt"
                        ]
                      }
                    }
                  },
                  "required": [
                    "results"
                  ]
                }
              }
            }
          }
        }
      }
    },
    "/api/queries": {
      "get": {
        "summary": "List all queries",
        "tags": [
          "control"
        ],
        "parameters": [
          {
            "schema": {
              "type": "string",
              "enum": [
                "cashConverters",
                "ebay",
                "gumtree",
                "salvos",
                "csMarket",
                "steamMarket",
                "csTradeBot"
              ]
            },
            "in": "query",
            "name": "type",
            "required": false
          }
        ],
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "queries": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "additionalProperties": false,
                        "properties": {
                          "type": {
                            "type": "string",
                            "enum": [
                              "cashConverters",
                              "ebay",
                              "gumtree",
                              "salvos",
                              "csMarket",
                              "steamMarket",
                              "csTradeBot"
                            ]
                          },
                          "id": {
                            "type": "string"
                          },
                          "dmOnly": {
                            "type": "boolean"
                          },
                          "url": {
                            "type": "string"
                          },
                          "name": {
                            "type": "string"
                          },
                          "displayUrl": {
                            "type": "string"
                          },
                          "maxPrice": {
                            "type": "number"
                          },
                          "minPrice": {
                            "type": "number"
                          },
                          "minFloat": {
                            "type": "number"
                          },
                          "maxFloat": {
                            "type": "number"
                          },
                          "requiredPhrases": {
                            "type": "string"
                          },
                          "excludePhrases": {
                            "type": "string"
                          },
                          "scanMode": {
                            "type": "string",
                            "enum": [
                              "searchUrl",
                              "siteWide"
                            ]
                          }
                        },
                        "required": [
                          "type",
                          "id",
                          "dmOnly"
                        ]
                      }
                    }
                  },
                  "required": [
                    "queries"
                  ]
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a query",
        "tags": [
          "control"
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "additionalProperties": false,
                "properties": {
                  "type": {
                    "type": "string",
                    "enum": [
                      "cashConverters",
                      "ebay",
                      "gumtree",
                      "salvos",
                      "csMarket",
                      "steamMarket",
                      "csTradeBot"
                    ]
                  },
                  "payload": {
                    "type": "object",
                    "additionalProperties": true
                  }
                },
                "required": [
                  "type",
                  "payload"
                ]
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "success": {
                      "type": "boolean"
                    },
                    "query": {
                      "type": "object",
                      "additionalProperties": true
                    }
                  },
                  "required": [
                    "success",
                    "query"
                  ]
                }
              }
            }
          }
        }
      },
      "put": {
        "summary": "Update a query",
        "tags": [
          "control"
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "additionalProperties": false,
                "properties": {
                  "type": {
                    "type": "string",
                    "enum": [
                      "cashConverters",
                      "ebay",
                      "gumtree",
                      "salvos",
                      "csMarket",
                      "steamMarket",
                      "csTradeBot"
                    ]
                  },
                  "id": {
                    "type": "string"
                  },
                  "payload": {
                    "type": "object",
                    "additionalProperties": true
                  }
                },
                "required": [
                  "type",
                  "id",
                  "payload"
                ]
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "success": {
                      "type": "boolean"
                    },
                    "query": {
                      "type": "object",
                      "additionalProperties": true
                    }
                  },
                  "required": [
                    "success",
                    "query"
                  ]
                }
              }
            }
          }
        }
      },
      "delete": {
        "summary": "Delete a query",
        "tags": [
          "control"
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "additionalProperties": false,
                "properties": {
                  "type": {
                    "type": "string",
                    "enum": [
                      "cashConverters",
                      "ebay",
                      "gumtree",
                      "salvos",
                      "csMarket",
                      "steamMarket",
                      "csTradeBot"
                    ]
                  },
                  "id": {
                    "type": "string"
                  }
                },
                "required": [
                  "type",
                  "id"
                ]
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "success": {
                      "type": "boolean"
                    }
                  },
                  "required": [
                    "success"
                  ]
                }
              }
            }
          }
        }
      }
    },
    "/api/testing/capabilities": {
      "get": {
        "summary": "Get testing API capabilities",
        "tags": [
          "testing"
        ],
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "testingEnabled": {
                      "type": "boolean"
                    },
                    "notificationProviders": {
                      "type": "array",
                      "items": {
                        "type": "string"
                      }
                    },
                    "discordConnected": {
                      "type": "boolean"
                    },
                    "availableQueryTypes": {
                      "type": "array",
                      "items": {
                        "type": "string"
                      }
                    },
                    "availableScannerTypes": {
                      "type": "array",
                      "items": {
                        "type": "string"
                      }
                    }
                  },
                  "required": [
                    "testingEnabled",
                    "notificationProviders",
                    "discordConnected",
                    "availableQueryTypes",
                    "availableScannerTypes"
                  ]
                }
              }
            }
          }
        }
      }
    },
    "/api/testing/notifications": {
      "get": {
        "summary": "List recent test notification history",
        "tags": [
          "testing"
        ],
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "results": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "additionalProperties": false,
                        "properties": {
                          "id": {
                            "type": "string"
                          },
                          "startedAt": {
                            "type": "string"
                          },
                          "finishedAt": {
                            "type": "string"
                          },
                          "durationMs": {
                            "type": "number"
                          },
                          "outcomes": {
                            "type": "array",
                            "items": {
                              "type": "object",
                              "additionalProperties": false,
                              "properties": {
                                "provider": {
                                  "type": "string"
                                },
                                "status": {
                                  "type": "string",
                                  "enum": [
                                    "sent",
                                    "failed"
                                  ]
                                },
                                "reason": {
                                  "type": [
                                    "string",
                                    "null"
                                  ]
                                }
                              },
                              "required": [
                                "provider",
                                "status"
                              ]
                            }
                          },
                          "error": {
                            "type": [
                              "string",
                              "null"
                            ]
                          }
                        },
                        "required": [
                          "id",
                          "startedAt",
                          "finishedAt",
                          "durationMs",
                          "outcomes",
                          "error"
                        ]
                      }
                    }
                  },
                  "required": [
                    "results"
                  ]
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Send a test notification through NotificationService",
        "tags": [
          "testing"
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "additionalProperties": false,
                "properties": {
                  "kind": {
                    "type": "string",
                    "enum": [
                      "deal",
                      "error"
                    ]
                  },
                  "source": {
                    "type": "string"
                  },
                  "title": {
                    "type": "string"
                  },
                  "url": {
                    "type": "string"
                  },
                  "price": {
                    "type": "number"
                  },
                  "imageUrl": {
                    "type": "string"
                  },
                  "message": {
                    "type": "string"
                  },
                  "query": {
                    "type": "object",
                    "additionalProperties": false,
                    "properties": {
                      "type": {
                        "type": "string"
                      },
                      "id": {
                        "type": "string"
                      }
                    },
                    "required": [
                      "type",
                      "id"
                    ]
                  },
                  "tags": {
                    "type": "array",
                    "items": {
                      "type": "string"
                    }
                  }
                },
                "required": [
                  "kind"
                ]
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "id": {
                      "type": "string"
                    },
                    "startedAt": {
                      "type": "string"
                    },
                    "finishedAt": {
                      "type": "string"
                    },
                    "durationMs": {
                      "type": "number"
                    },
                    "outcomes": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "additionalProperties": false,
                        "properties": {
                          "provider": {
                            "type": "string"
                          },
                          "status": {
                            "type": "string",
                            "enum": [
                              "sent",
                              "failed"
                            ]
                          },
                          "reason": {
                            "type": [
                              "string",
                              "null"
                            ]
                          }
                        },
                        "required": [
                          "provider",
                          "status"
                        ]
                      }
                    },
                    "error": {
                      "type": [
                        "string",
                        "null"
                      ]
                    }
                  },
                  "required": [
                    "id",
                    "startedAt",
                    "finishedAt",
                    "durationMs",
                    "outcomes",
                    "error"
                  ]
                }
              }
            }
          }
        }
      }
    },
    "/api/testing/scans/run": {
      "post": {
        "summary": "Run a temporary scan without persisting a user query",
        "tags": [
          "testing"
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "additionalProperties": false,
                "properties": {
                  "type": {
                    "type": "string",
                    "enum": [
                      "cashConverters",
                      "ebay",
                      "gumtree",
                      "salvos",
                      "csMarket",
                      "steamMarket",
                      "csTradeBot"
                    ]
                  },
                  "payload": {
                    "type": "object",
                    "additionalProperties": true
                  },
                  "notify": {
                    "type": "boolean"
                  }
                },
                "required": [
                  "type",
                  "payload"
                ]
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "id": {
                      "type": "string"
                    },
                    "startedAt": {
                      "type": "string"
                    },
                    "finishedAt": {
                      "type": "string"
                    },
                    "durationMs": {
                      "type": "number"
                    },
                    "items": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "additionalProperties": false,
                        "properties": {
                          "source": {
                            "type": "string"
                          },
                          "title": {
                            "type": "string"
                          },
                          "url": {
                            "type": "string"
                          },
                          "price": {
                            "type": [
                              "number",
                              "null"
                            ]
                          },
                          "imageUrl": {
                            "type": [
                              "string",
                              "null"
                            ]
                          },
                          "passedFilters": {
                            "type": "boolean"
                          },
                          "filterReason": {
                            "type": [
                              "string",
                              "null"
                            ]
                          }
                        },
                        "required": [
                          "source",
                          "title",
                          "url",
                          "price",
                          "imageUrl",
                          "passedFilters",
                          "filterReason"
                        ]
                      }
                    },
                    "notifications": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "additionalProperties": false,
                        "properties": {
                          "source": {
                            "type": "string"
                          },
                          "title": {
                            "type": "string"
                          },
                          "url": {
                            "type": "string"
                          },
                          "price": {
                            "type": "number"
                          },
                          "imageUrl": {
                            "type": "string"
                          }
                        },
                        "required": [
                          "source",
                          "title",
                          "url"
                        ]
                      }
                    },
                    "errors": {
                      "type": "array",
                      "items": {
                        "type": "string"
                      }
                    },
                    "notificationsPublished": {
                      "type": "boolean"
                    }
                  },
                  "required": [
                    "id",
                    "startedAt",
                    "finishedAt",
                    "durationMs",
                    "items",
                    "notifications",
                    "errors",
                    "notificationsPublished"
                  ]
                }
              }
            }
          }
        }
      }
    },
    "/api/testing/runs": {
      "get": {
        "summary": "List recent temporary scan run history",
        "tags": [
          "testing"
        ],
        "responses": {
          "200": {
            "description": "Default Response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "additionalProperties": false,
                  "properties": {
                    "results": {
                      "type": "array",
                      "items": {
                        "type": "object",
                        "additionalProperties": false,
                        "properties": {
                          "id": {
                            "type": "string"
                          },
                          "startedAt": {
                            "type": "string"
                          },
                          "finishedAt": {
                            "type": "string"
                          },
                          "durationMs": {
                            "type": "number"
                          },
                          "items": {
                            "type": "array",
                            "items": {
                              "type": "object",
                              "additionalProperties": false,
                              "properties": {
                                "source": {
                                  "type": "string"
                                },
                                "title": {
                                  "type": "string"
                                },
                                "url": {
                                  "type": "string"
                                },
                                "price": {
                                  "type": [
                                    "number",
                                    "null"
                                  ]
                                },
                                "imageUrl": {
                                  "type": [
                                    "string",
                                    "null"
                                  ]
                                },
                                "passedFilters": {
                                  "type": "boolean"
                                },
                                "filterReason": {
                                  "type": [
                                    "string",
                                    "null"
                                  ]
                                }
                              },
                              "required": [
                                "source",
                                "title",
                                "url",
                                "price",
                                "imageUrl",
                                "passedFilters",
                                "filterReason"
                              ]
                            }
                          },
                          "notifications": {
                            "type": "array",
                            "items": {
                              "type": "object",
                              "additionalProperties": false,
                              "properties": {
                                "source": {
                                  "type": "string"
                                },
                                "title": {
                                  "type": "string"
                                },
                                "url": {
                                  "type": "string"
                                },
                                "price": {
                                  "type": "number"
                                },
                                "imageUrl": {
                                  "type": "string"
                                }
                              },
                              "required": [
                                "source",
                                "title",
                                "url"
                              ]
                            }
                          },
                          "errors": {
                            "type": "array",
                            "items": {
                              "type": "string"
                            }
                          },
                          "notificationsPublished": {
                            "type": "boolean"
                          }
                        },
                        "required": [
                          "id",
                          "startedAt",
                          "finishedAt",
                          "durationMs",
                          "items",
                          "notifications",
                          "errors",
                          "notificationsPublished"
                        ]
                      }
                    }
                  },
                  "required": [
                    "results"
                  ]
                }
              }
            }
          }
        }
      }
    }
  }
}`

func main() {
	log.Println("Generating OpenAPI spec...")

	apiDir := filepath.Join("..", "frontend", "src", "api")
	openApiPath := filepath.Join(apiDir, "openapi.json")

	// Ensure API dir exists
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		log.Fatalf("Failed to create frontend API directory: %v", err)
	}

	// Write openapi.json
	if err := os.WriteFile(openApiPath, []byte(openApiJSON+"\n"), 0644); err != nil {
		log.Fatalf("Failed to write openapi.json: %v", err)
	}
	log.Printf("Successfully wrote openapi.json to %s", openApiPath)

	// Run npx openapi-typescript to generate schema.d.ts
	log.Println("Regenerating schema.d.ts typings...")
	cmd := exec.Command("npx", "openapi-typescript", openApiPath, "-o", filepath.Join(apiDir, "schema.d.ts"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Failed to execute openapi-typescript: %v", err)
	}

	log.Println("OpenAPI schema typings regenerated successfully.")
}
