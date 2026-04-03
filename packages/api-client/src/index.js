import createClient from "openapi-fetch";
export const client = createClient({
    baseUrl: typeof window === "undefined" ? "" : window.__VERIN_API__,
    credentials: "include",
});
async function unwrap(promise) {
    const response = await promise;
    if (response.error) {
        throw response.error;
    }
    return response.data;
}
export const api = {
    login(payload) {
        return unwrap(client.POST("/api/v1/auth/login", { body: payload }));
    },
    logout() {
        return client.POST("/api/v1/auth/logout");
    },
    me() {
        return unwrap(client.GET("/api/v1/auth/me"));
    },
    setupMFA() {
        return unwrap(client.POST("/api/v1/auth/mfa/setup"));
    },
    verifyMFA(payload) {
        return unwrap(client.POST("/api/v1/auth/mfa/verify", { body: payload }));
    },
    listDocuments(search = "") {
        return unwrap(client.GET("/api/v1/documents", {
            params: { query: { q: search } },
        }));
    },
    getDocument(documentId) {
        return unwrap(client.GET("/api/v1/documents/{documentId}", { params: { path: { documentId } } }));
    },
    initUpload(payload) {
        return unwrap(client.POST("/api/v1/documents/init-upload", { body: payload }));
    },
    completeUpload(payload) {
        return unwrap(client.POST("/api/v1/documents/complete-upload", { body: payload }));
    },
    search(query) {
        return unwrap(client.GET("/api/v1/search", { params: { query: { q: query } } }));
    },
    advancedSearch(payload) {
        return unwrap(client.POST("/api/v1/search/advanced", { body: payload }));
    },
    savedSearches() {
        return unwrap(client.GET("/api/v1/search/saved"));
    },
    createSavedSearch(payload) {
        return unwrap(client.POST("/api/v1/search/saved", { body: payload }));
    },
    auditEvents() {
        return unwrap(client.GET("/api/v1/audit/events"));
    },
    exportAudit() {
        return unwrap(client.POST("/api/v1/audit/reports/export"));
    },
    notifications() {
        return unwrap(client.GET("/api/v1/notifications"));
    },
    markNotificationRead(notificationId) {
        return client.POST("/api/v1/notifications/{notificationId}/read", { params: { path: { notificationId } } });
    },
    adminUsers() {
        return unwrap(client.GET("/api/v1/admin/users"));
    },
    adminRoles() {
        return unwrap(client.GET("/api/v1/admin/roles"));
    },
    adminQuotas() {
        return unwrap(client.GET("/api/v1/admin/quotas"));
    },
    upsertQuota(payload) {
        return unwrap(client.POST("/api/v1/admin/quotas", { body: payload }));
    },
    adminRetention() {
        return unwrap(client.GET("/api/v1/admin/retention"));
    },
    upsertRetention(payload) {
        return unwrap(client.POST("/api/v1/admin/retention", { body: payload }));
    },
    adminSettings() {
        return unwrap(client.GET("/api/v1/admin/settings"));
    },
    upsertSetting(payload) {
        return unwrap(client.POST("/api/v1/admin/settings", { body: payload }));
    },
    adminJobs() {
        return unwrap(client.GET("/api/v1/admin/jobs"));
    },
    adminUsage() {
        return unwrap(client.GET("/api/v1/admin/usage"));
    },
    adminHealth() {
        return unwrap(client.GET("/api/v1/admin/health"));
    },
};
