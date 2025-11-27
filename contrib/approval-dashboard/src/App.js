import React, { useEffect, useState } from "react";
import axios from "axios";
import ReactDiffViewer from "react-diff-viewer-continued";

const api = axios.create({
  baseURL: "http://localhost:8080",
});

function App() {
  const [requests, setRequests] = useState([]);
  const [selected, setSelected] = useState(null);
  const [logs, setLogs] = useState([]);
  const [toast, setToast] = useState(null);

  // 🔐 RBAC ile ilgili state
  const [selectedUser, setSelectedUser] = useState("yusuf"); // aktif kullanıcı
  const [currentUser, setCurrentUser] = useState(null); // /me'den gelen user
  const [userLoading, setUserLoading] = useState(false);

  // Toast helper
  const showToast = (msg) => {
    setToast(msg);
    setTimeout(() => setToast(null), 2500);
  };

  // Seçili kullanıcı değiştiğinde: önce /me, sonra request listesi
  useEffect(() => {
    fetchCurrentUser();
    fetchRequests();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedUser]);

  // --- Aktif kullanıcı bilgisini backend'den çek (/me) ---
  const fetchCurrentUser = async () => {
    setUserLoading(true);
    try {
      const res = await api.get("/me", {
        headers: {
          "X-User": selectedUser,
        },
      });
      setCurrentUser(res.data);
    } catch (err) {
      console.error("Kullanıcı bilgisi alınamadı:", err);
      setCurrentUser(null);
      showToast("⚠️ Kullanıcı bilgisi alınamadı (X-User hatası olabilir)");
    } finally {
      setUserLoading(false);
    }
  };

  // --- Tüm approval request'leri getir ---
  const fetchRequests = async () => {
    try {
      const res = await api.get("/approval-requests", {
        headers: {
          "X-User": selectedUser,
        },
      });
      setRequests(res.data);
    } catch (err) {
      console.error("Veri çekilemedi:", err);
      showToast("⚠️ Veriler alınamadı");
    }
  };

  // --- Approve / Reject işlemleri (sadece ADMIN başarılı olacak) ---
  const handleAction = async (id, action) => {
    try {
      await api.post(
        `/approval-requests/${id}/${action}`,
        {},
        {
          headers: {
            "X-User": selectedUser,
          },
        }
      );

      showToast(
        action === "approve" ? "✔️ Request approved" : "✖️ Request rejected"
      );

      // Listeyi yenile
      fetchRequests();

      // Detay paneli açıksa status'u güncelle
      if (selected && selected.id === id) {
        const updated = requests.find((r) => r.id === id);
        if (updated) {
          setSelected({
            ...updated,
            status: action === "approve" ? "APPROVED" : "REJECTED",
          });
        }
      }
    } catch (err) {
      console.error("İşlem başarısız:", err);
      if (err.response && err.response.status === 403) {
        showToast("⛔ Yetkin yok (ADMIN değil)");
      } else if (err.response && err.response.status === 401) {
        showToast("🔒 Yetkisiz: X-User veya kullanıcı hatalı");
      } else {
        showToast("⚠️ İşlem başarısız");
      }
    }
  };

  // --- Detay (View) butonu ---
  const handleView = async (req) => {
    setSelected(req);
    try {
      const res = await api.get(`/approval-logs/${req.id}`, {
        headers: {
          "X-User": selectedUser,
        },
      });
      setLogs(res.data);
    } catch (err) {
      console.error("Loglar alınamadı:", err);
      setLogs([]);
      showToast("⚠️ Loglar alınamadı");
    }
  };

  // --- Kullanıcı rolüne göre approve/reject gösterilsin mi? ---
  const canApproveReject =
    currentUser && currentUser.role && currentUser.role.toUpperCase() === "ADMIN";

  // Status'e göre renkli badge
  const renderStatusBadge = (status) => {
    const base = {
      padding: "2px 8px",
      borderRadius: "999px",
      fontSize: "12px",
      fontWeight: "bold",
    };

    if (status === "PENDING")
      return <span style={{ ...base, background: "#fff3cd", color: "#856404" }}>PENDING</span>;
    if (status === "APPROVED")
      return <span style={{ ...base, background: "#d4edda", color: "#155724" }}>APPROVED</span>;
    if (status === "REJECTED")
      return <span style={{ ...base, background: "#f8d7da", color: "#721c24" }}>REJECTED</span>;
    return <span style={base}>{status}</span>;
  };

  return (
    <div style={{ padding: "30px", fontFamily: "Arial, sans-serif" }}>
      {/* 🔔 Toast */}
      {toast && (
        <div
          style={{
            position: "fixed",
            top: 16,
            right: 16,
            background: "#333",
            color: "#fff",
            padding: "10px 14px",
            borderRadius: 8,
            zIndex: 9999,
            boxShadow: "0 6px 16px rgba(0,0,0,.25)",
            fontSize: "14px",
          }}
        >
          {toast}
        </div>
      )}

      <h2>🚀 Flipt Approval Dashboard (RBAC + Audit + Diff)</h2>

      {/* 🔐 Aktif Kullanıcı Seçimi + Bilgi */}
      <div
        style={{
          marginTop: "10px",
          marginBottom: "20px",
          padding: "10px 15px",
          borderRadius: "8px",
          border: "1px solid #ddd",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: "16px",
        }}
      >
        <div>
          <label>
            <b>Aktif Kullanıcı: </b>
            <select
              value={selectedUser}
              onChange={(e) => setSelectedUser(e.target.value)}
              style={{ marginLeft: 8, padding: "4px 8px" }}
            >
              <option value="yusuf">yusuf (ADMIN)</option>
              <option value="dev1">dev1 (DEVELOPER)</option>
              <option value="viewer1">viewer1 (VIEWER)</option>
            </select>
          </label>
        </div>

        <div style={{ fontSize: "14px", textAlign: "right" }}>
          {userLoading ? (
            <span>🔄 Kullanıcı yükleniyor...</span>
          ) : currentUser ? (
            <>
              <div>
                <b>{currentUser.full_name || currentUser.username}</b>
              </div>
              <div>
                Rol: <span style={{ fontWeight: "bold" }}>{currentUser.role}</span>
              </div>
            </>
          ) : (
            <span>⚠️ Kullanıcı bilgisi alınamadı</span>
          )}
        </div>
      </div>

      {/* --- Request Tablosu --- */}
      <table
        border="1"
        cellPadding="8"
        cellSpacing="0"
        width="100%"
        style={{ borderCollapse: "collapse", marginTop: "10px" }}
      >
        <thead>
          <tr style={{ background: "#f0f0f0" }}>
            <th>ID</th>
            <th>Source</th>
            <th>Target</th>
            <th>Status</th>
            <th>Requested By</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {requests.length > 0 ? (
            requests.map((req) => (
              <tr key={req.id}>
                <td>{String(req.id).slice(0, 8)}...</td>
                <td>{req.source_env}</td>
                <td>{req.target_env}</td>
                <td>{renderStatusBadge(req.status)}</td>
                <td>{req.requested_by}</td>
                <td>
                  <button
                    onClick={() => handleView(req)}
                    style={{
                      marginRight: "10px",
                      backgroundColor: "#ddd",
                      border: "none",
                      padding: "5px 10px",
                      cursor: "pointer",
                    }}
                  >
                    🔍 View
                  </button>

                  {req.status === "PENDING" ? (
                    canApproveReject ? (
                      <>
                        <button
                          onClick={() => handleAction(req.id, "approve")}
                          style={{
                            marginRight: "10px",
                            backgroundColor: "#4caf50",
                            color: "white",
                            border: "none",
                            padding: "5px 10px",
                            cursor: "pointer",
                          }}
                        >
                          ✅ Approve
                        </button>
                        <button
                          onClick={() => handleAction(req.id, "reject")}
                          style={{
                            backgroundColor: "#f44336",
                            color: "white",
                            border: "none",
                            padding: "5px 10px",
                            cursor: "pointer",
                          }}
                        >
                          ❌ Reject
                        </button>
                      </>
                    ) : (
                      <em>Only ADMIN can approve/reject</em>
                    )
                  ) : (
                    <em>{req.status}</em>
                  )}
                </td>
              </tr>
            ))
          ) : (
            <tr>
              <td colSpan="6" style={{ textAlign: "center", padding: "20px" }}>
                No requests found
              </td>
            </tr>
          )}
        </tbody>
      </table>

      {/* --- Detay Paneli --- */}
      {selected && (
        <div
          style={{
            marginTop: "30px",
            padding: "20px",
            border: "1px solid #ccc",
            borderRadius: "8px",
            backgroundColor: "#fafafa",
          }}
        >
          <h3>Request Details</h3>
          <p>
            <b>Source:</b> {selected.source_env}
          </p>
          <p>
            <b>Target:</b> {selected.target_env}
          </p>
          <p>
            <b>Status:</b> {selected.status}</p>
          <p>
            <b>Requested By:</b> {selected.requested_by}
          </p>

          {/* --- Change Payload: Diff Görünümü --- */}
          <h4>Change Payload (Diff View)</h4>

          {(() => {
            let payload = selected.change_payload;
            let oldData = {};
            let newData = {};

            if (typeof payload === "string") {
              try {
                payload = JSON.parse(payload);
              } catch (e) {
                console.error("Payload JSON değil:", e);
              }
            }

            if (payload && payload.old && payload.new) {
              oldData = payload.old;
              newData = payload.new;
            } else {
              oldData = {};
              newData = payload || {};
            }

            return (
              <div
                style={{
                  background: "#f9f9f9",
                  borderRadius: "6px",
                  padding: "10px",
                  marginBottom: "20px",
                }}
              >
                <ReactDiffViewer
                  oldValue={JSON.stringify(oldData, null, 2)}
                  newValue={JSON.stringify(newData, null, 2)}
                  splitView={true}
                  showDiffOnly={true}
                  leftTitle="Before"
                  rightTitle="After"
                />
              </div>
            );
          })()}

          {/* --- Approval Logs --- */}
          <h4>Approval Logs</h4>
          {Array.isArray(logs) && logs.length > 0 ? (
            <ul>
              {logs.map((log, index) => (
                <li key={index}>
                  [{new Date(log.timestamp).toLocaleString()}]{" "}
                  <b>{log.actor}</b> → {log.action}
                </li>
              ))}
            </ul>
          ) : (
            <em>No logs yet</em>
          )}
        </div>
      )}
    </div>
  );
}

export default App;
