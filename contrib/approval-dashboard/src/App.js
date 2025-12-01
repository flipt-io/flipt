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

  // RBAC-related state
  const [selectedUser, setSelectedUser] = useState("yusuf"); // active user
  const [currentUser, setCurrentUser] = useState(null); // user from /me
  const [userLoading, setUserLoading] = useState(false);

  // Toast helper
  const showToast = (msg) => {
    setToast(msg);
    setTimeout(() => setToast(null), 2500);
  };

  // When selected user changes: fetch /me first, then request list
  useEffect(() => {
    fetchCurrentUser();
    fetchRequests();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedUser]);

  // --- Fetch current user from backend (/me) ---
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
      console.error("Could not fetch user info:", err);
      setCurrentUser(null);
      showToast("⚠️ Could not fetch user info (possible X-User issue)");
    } finally {
      setUserLoading(false);
    }
  };

  // --- Fetch all approval requests ---
  const fetchRequests = async () => {
    try {
      // Backend: GET /api/v1/changes
      const res = await api.get("/api/v1/changes", {
        headers: {
          "X-User": selectedUser,
        },
      });
      setRequests(res.data);
    } catch (err) {
      console.error("Could not fetch data:", err);
      showToast("⚠️ Could not fetch data");
    }
  };

  // --- Approve / Reject actions (only ADMIN allowed) ---
  const handleAction = async (id, action) => {
    try {
      await api.post(
        `/api/v1/changes/${id}/${action}`,
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

      // Refresh list
      fetchRequests();

      // Update status if detail panel is open
      if (selected && selected.id === id) {
        const updated = requests.find((r) => r.id === id);
        if (updated) {
          setSelected({
            ...updated,
            status: action === "approve" ? "APPROVED" : "REJECTED",
          });
        } else {
          setSelected({
            ...selected,
            status: action === "approve" ? "APPROVED" : "REJECTED",
          });
        }
      }
    } catch (err) {
      console.error("Action failed:", err);
      if (err.response && err.response.status === 403) {
        showToast("⛔ Forbidden (not ADMIN)");
      } else if (err.response && err.response.status === 401) {
        showToast("🔒 Unauthorized: X-User missing or invalid");
      } else if (err.response && err.response.status === 400) {
        // Örn: "cannot approve before review"
        showToast(`⚠️ ${err.response.data || "Bad Request"}`);
      } else {
        showToast("⚠️ Action failed");
      }
    }
  };

  // --- Multi-level Review action (DEVELOPER or ADMIN) ---
  const handleReview = async (id) => {
    try {
      await api.post(
        `/api/v1/changes/${id}/review`,
        {},
        {
          headers: {
            "X-User": selectedUser,
          },
        }
      );

      showToast("👀 Request reviewed");

      // Refresh list
      fetchRequests();

      // Update detail panel if open
      if (selected && selected.id === id) {
        setSelected({
          ...selected,
          review_state: "REVIEWED",
        });
      }
    } catch (err) {
      console.error("Review failed:", err);
      if (err.response && err.response.status === 403) {
        showToast("⛔ Forbidden (not REVIEWER/ADMIN)");
      } else if (err.response && err.response.status === 401) {
        showToast("🔒 Unauthorized");
      } else {
        showToast("⚠️ Review failed");
      }
    }
  };

  // --- Detail (View) button ---
  const handleView = async (req) => {
    setSelected(req);
    try {
      const res = await api.get(`/api/v1/changes/${req.id}/logs`, {
        headers: {
          "X-User": selectedUser,
        },
      });
      setLogs(res.data);
    } catch (err) {
      console.error("Could not fetch logs:", err);
      setLogs([]);
      showToast("⚠️ Could not fetch logs");
    }
  };

  // --- Permissions helpers ---
  const canApproveReject =
    currentUser &&
    currentUser.role &&
    currentUser.role.toUpperCase() === "ADMIN";

  const canReview =
    currentUser &&
    currentUser.role &&
    ["DEVELOPER", "ADMIN"].includes(currentUser.role.toUpperCase());

  // Colored badge based on status
  const renderStatusBadge = (status) => {
    const base = {
      padding: "2px 8px",
      borderRadius: "999px",
      fontSize: "12px",
      fontWeight: "bold",
    };

    if (status === "PENDING")
      return (
        <span style={{ ...base, background: "#fff3cd", color: "#856404" }}>
          PENDING
        </span>
      );
    if (status === "APPROVED")
      return (
        <span style={{ ...base, background: "#d4edda", color: "#155724" }}>
          APPROVED
        </span>
      );
    if (status === "REJECTED")
      return (
        <span style={{ ...base, background: "#f8d7da", color: "#721c24" }}>
          REJECTED
        </span>
      );
    return <span style={base}>{status}</span>;
  };

  const renderReviewBadge = (state) => {
    const base = {
      padding: "2px 8px",
      borderRadius: "999px",
      fontSize: "12px",
      fontWeight: "bold",
    };
    const value = (state || "NONE").toUpperCase();

    if (value === "REVIEWED") {
      return (
        <span style={{ ...base, background: "#d1ecf1", color: "#0c5460" }}>
          REVIEWED
        </span>
      );
    }

    return (
      <span style={{ ...base, background: "#e2e3e5", color: "#383d41" }}>
        NONE
      </span>
    );
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

      <h2>🚀 Flipt Approval Dashboard (RBAC + Audit + Diff + Multi-level)</h2>

      {/* Active User Selection + Info */}
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
            <b>Active User: </b>
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
            <span>🔄 Loading user...</span>
          ) : currentUser ? (
            <>
              <div>
                <b>{currentUser.full_name || currentUser.username}</b>
              </div>
              <div>
                Role:{" "}
                <span style={{ fontWeight: "bold" }}>{currentUser.role}</span>
              </div>
            </>
          ) : (
            <span>⚠️ Could not fetch user info</span>
          )}
        </div>
      </div>

      {/* --- Request Table --- */}
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
            <th>Source Env</th>
            <th>Target Env</th>
            <th>Source Branch</th>
            <th>Target Branch</th>
            <th>Repo</th>
            <th>Review</th>
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
                <td>{req.source_branch || "-"}</td>
                <td>{req.target_branch || "-"} </td>
                <td>{req.repo_url || "-"}</td>
                <td>{renderReviewBadge(req.review_state)}</td>
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

                  {/* REVIEW BUTTON */}
                  {req.status === "PENDING" &&
                    req.review_state !== "REVIEWED" &&
                    canReview && (
                      <button
                        onClick={() => handleReview(req.id)}
                        style={{
                          marginRight: "10px",
                          backgroundColor: "#2196f3",
                          color: "white",
                          border: "none",
                          padding: "5px 10px",
                          cursor: "pointer",
                        }}
                      >
                        👀 Review
                      </button>
                    )}

                  {/* APPROVE / REJECT */}
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
              <td colSpan="10" style={{ textAlign: "center", padding: "20px" }}>
                No requests found
              </td>
            </tr>
          )}
        </tbody>
      </table>

      {/* --- Detail Panel --- */}
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
            <b>ID:</b> {selected.id}
          </p>
          <p>
            <b>Source Env:</b> {selected.source_env}
          </p>
          <p>
            <b>Target Env:</b> {selected.target_env}
          </p>
          <p>
            <b>Source Branch:</b> {selected.source_branch || "-"}
          </p>
          <p>
            <b>Target Branch:</b> {selected.target_branch || "-"}
          </p>
          <p>
            <b>Repo URL:</b> {selected.repo_url || "-"}
          </p>
          <p>
            <b>Review State:</b> {selected.review_state || "NONE"}
          </p>
          <p>
            <b>Status:</b> {selected.status}
          </p>
          <p>
            <b>Requested By:</b> {selected.requested_by}
          </p>

          {/* --- Change Payload: Diff View --- */}
          <h4>Change Payload (Diff View)</h4>

          {(() => {
            let payload = selected.change_payload;
            let oldData = {};
            let newData = {};

            if (typeof payload === "string") {
              try {
                payload = JSON.parse(payload);
              } catch (e) {
                console.error("Payload not JSON:", e);
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
