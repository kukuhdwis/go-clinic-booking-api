/**
 * ClinicEase — Clinical Care & Patient Scheduling
 * Standard business interface communicating with Go REST API
 */

// API Configuration
const API_BASE = window.location.protocol === 'file:' 
  ? 'http://localhost:8080/api/v1' 
  : `${window.location.origin}/api/v1`;

// State
let state = {
  token: localStorage.getItem('clinicease_token') || null,
  user: JSON.parse(localStorage.getItem('clinicease_user') || 'null'),
  doctors: [],
  selectedDoctor: null,
  selectedDate: '',
  selectedSlot: null,
  appointments: []
};

// DOM Elements
const elements = {
  authLoggedIn: document.getElementById('auth-logged-in'),
  authLoggedOut: document.getElementById('auth-logged-out'),
  currentUserName: document.getElementById('current-user-name'),
  currentUserRole: document.getElementById('current-user-role'),
  userAvatarInitials: document.getElementById('user-avatar-initials'),
  btnLogout: document.getElementById('btn-logout'),
  btnOpenLogin: document.getElementById('btn-open-login'),
  btnOpenRegister: document.getElementById('btn-open-register'),
  btnQuickPatient: document.getElementById('btn-quick-patient'),
  btnQuickDoctor: document.getElementById('btn-quick-doctor'),
  btnQuickAdmin: document.getElementById('btn-quick-admin'),
  modalAuth: document.getElementById('modal-auth'),
  modalAuthTitle: document.getElementById('modal-auth-title'),
  formAuth: document.getElementById('form-auth'),
  btnCloseAuth: document.getElementById('btn-close-auth'),
  btnToggleAuthMode: document.getElementById('btn-toggle-auth-mode'),
  groupFullname: document.getElementById('group-fullname'),
  groupPhone: document.getElementById('group-phone'),
  authFullname: document.getElementById('auth-fullname'),
  authPhone: document.getElementById('auth-phone'),
  authEmail: document.getElementById('auth-email'),
  authPassword: document.getElementById('auth-password'),
  btnSubmitAuth: document.getElementById('btn-submit-auth'),
  doctorsList: document.getElementById('doctors-list'),
  bookingDate: document.getElementById('booking-date'),
  btnDateTomorrow: document.getElementById('btn-date-tomorrow'),
  btnDateMonday: document.getElementById('btn-date-monday'),
  slotsGrid: document.getElementById('slots-grid'),
  sumDoctorName: document.getElementById('sum-doctor-name'),
  sumSpecialization: document.getElementById('sum-specialization'),
  sumDate: document.getElementById('sum-date'),
  sumSlot: document.getElementById('sum-slot'),
  sumFee: document.getElementById('sum-fee'),
  complaintText: document.getElementById('complaint-text'),
  btnConfirmBooking: document.getElementById('btn-confirm-booking'),
  badgeApptCount: document.getElementById('badge-appt-count'),
  appointmentsList: document.getElementById('appointments-list'),
  filterStatus: document.getElementById('filter-status'),
  btnRefreshAppts: document.getElementById('btn-refresh-appts'),
  modalDiagnosis: document.getElementById('modal-diagnosis'),
  formDiagnosis: document.getElementById('form-diagnosis'),
  diagAppointmentId: document.getElementById('diag-appointment-id'),
  diagDiagnosis: document.getElementById('diag-diagnosis'),
  diagPrescription: document.getElementById('diag-prescription'),
  diagNotes: document.getElementById('diag-notes'),
  btnCloseDiagnosis: document.getElementById('btn-close-diagnosis'),
  btnCancelDiagnosis: document.getElementById('btn-cancel-diagnosis'),
  statDoctorsCount: document.getElementById('stat-doctors-count'),
  statBookingsCount: document.getElementById('stat-bookings-count'),
  btnOpenAddDoctor: document.getElementById('btn-open-add-doctor'),
  modalAddDoctor: document.getElementById('modal-add-doctor'),
  formAddDoctor: document.getElementById('form-add-doctor'),
  btnCloseAddDoctor: document.getElementById('btn-close-add-doctor'),
  btnCancelAddDoctor: document.getElementById('btn-cancel-add-doctor'),
  tabBtnPractice: document.getElementById('tab-btn-practice'),
  formDoctorProfile: document.getElementById('form-doctor-profile'),
  docProfileName: document.getElementById('doc-profile-name'),
  docProfileSpecialization: document.getElementById('doc-profile-specialization'),
  docProfileFee: document.getElementById('doc-profile-fee'),
  docProfileDuration: document.getElementById('doc-profile-duration'),
  docProfileBio: document.getElementById('doc-profile-bio'),
  doctorMySchedules: document.getElementById('doctor-my-schedules'),
  formAddSchedule: document.getElementById('form-add-schedule'),
  schedDay: document.getElementById('sched-day'),
  schedStart: document.getElementById('sched-start'),
  schedEnd: document.getElementById('sched-end')
};

let isRegisterMode = false;

// ==========================================================================
// API Request Helper
// ==========================================================================
async function apiRequest(endpoint, options = {}) {
  const url = `${API_BASE}${endpoint}`;
  const headers = {
    'Content-Type': 'application/json',
    ...(state.token ? { Authorization: `Bearer ${state.token}` } : {}),
    ...options.headers
  };

  try {
    const response = await fetch(url, { ...options, headers });
    let data = null;

    try {
      data = await response.json();
    } catch (e) {
      data = { message: response.statusText };
    }

    if (!response.ok) {
      throw {
        status: response.status,
        message: data.message || `HTTP ${response.status}`,
        errorCode: data.error_code || 'ERROR',
        errors: data.errors
      };
    }

    return data;
  } catch (err) {
    if (err.status) throw err;
    throw {
      status: 0,
      message: 'Unable to connect to the clinic server. Please check your network connection.',
      errorCode: 'OFFLINE'
    };
  }
}

// ==========================================================================
// Toast Notifications
// ==========================================================================
function showToast(message, type = 'info', duration = 4000) {
  const container = document.getElementById('toast-container');
  if (!container) return;

  const toast = document.createElement('div');

  let borderLeft = 'border-l-clinic-600 text-slate-800';
  let icon = 'ℹ️';

  if (type === 'success') {
    borderLeft = 'border-l-emerald-600 text-slate-900';
    icon = '✅';
  } else if (type === 'error') {
    borderLeft = 'border-l-rose-600 text-slate-900';
    icon = '⚠️';
  }

  toast.className = `p-3.5 rounded-xl shadow-lg border border-slate-200 bg-white text-xs font-medium flex items-center gap-3 pointer-events-auto transition-all transform duration-300 border-l-4 ${borderLeft}`;
  toast.innerHTML = `<span class="text-sm">${icon}</span> <span class="flex-1">${message}</span>`;
  container.appendChild(toast);

  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(20px)';
    setTimeout(() => toast.remove(), 300);
  }, duration);
}

// ==========================================================================
// Authentication & Profile State
// ==========================================================================
function updateAuthUI() {
  if (state.token && state.user) {
    elements.authLoggedOut.classList.add('hidden');
    elements.authLoggedIn.classList.remove('hidden');
    elements.currentUserName.textContent = state.user.full_name || state.user.email;
    elements.currentUserRole.textContent = state.user.role || 'patient';
    elements.userAvatarInitials.textContent = (state.user.full_name || 'U').charAt(0).toUpperCase();

    // Show "Add Specialist" button only for clinic administrators
    if (elements.btnOpenAddDoctor) {
      if (state.user.role === 'admin') {
        elements.btnOpenAddDoctor.classList.remove('hidden');
      } else {
        elements.btnOpenAddDoctor.classList.add('hidden');
      }
    }

    // Show "My Practice & Rates" tab only for doctors
    if (elements.tabBtnPractice) {
      if (state.user.role === 'doctor') {
        elements.tabBtnPractice.classList.remove('hidden');
      } else {
        elements.tabBtnPractice.classList.add('hidden');
        if (elements.tabBtnPractice.classList.contains('active')) {
          document.getElementById('tab-btn-book')?.click();
        }
      }
    }
  } else {
    elements.authLoggedIn.classList.add('hidden');
    elements.authLoggedOut.classList.remove('hidden');
    if (elements.btnOpenAddDoctor) {
      elements.btnOpenAddDoctor.classList.add('hidden');
    }
    if (elements.tabBtnPractice) {
      elements.tabBtnPractice.classList.add('hidden');
      if (elements.tabBtnPractice.classList.contains('active')) {
        document.getElementById('tab-btn-book')?.click();
      }
    }
  }
}

async function verifyProfile() {
  if (!state.token) return;
  try {
    const res = await apiRequest('/auth/me');
    state.user = res.data;
    localStorage.setItem('clinicease_user', JSON.stringify(res.data));
    updateAuthUI();
  } catch (err) {
    logout();
  }
}

function setSession(token, user) {
  state.token = token;
  state.user = user;
  localStorage.setItem('clinicease_token', token);
  localStorage.setItem('clinicease_user', JSON.stringify(user));
  updateAuthUI();
  fetchAppointments();
}

function logout() {
  state.token = null;
  state.user = null;
  localStorage.removeItem('clinicease_token');
  localStorage.removeItem('clinicease_user');
  updateAuthUI();
  showToast('You have signed out successfully.', 'info');
  fetchAppointments();
}

// ==========================================================================
// Tabs Navigation
// ==========================================================================
function initTabs() {
  const tabs = [
    { btnId: 'tab-btn-book', paneId: 'tab-book' },
    { btnId: 'tab-btn-appointments', paneId: 'tab-appointments' },
    { btnId: 'tab-btn-practice', paneId: 'tab-practice' }
  ];

  tabs.forEach(tab => {
    const btn = document.getElementById(tab.btnId);
    if (!btn) return;

    btn.addEventListener('click', () => {
      tabs.forEach(t => {
        const b = document.getElementById(t.btnId);
        const p = document.getElementById(t.paneId);
        if (b) b.className = 'nav-tab pb-4 px-1 border-b-2 font-medium text-sm flex items-center gap-2 border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300 transition';
        if (p) p.classList.add('hidden');
      });

      btn.className = 'nav-tab active pb-4 px-1 border-b-2 font-semibold text-sm flex items-center gap-2 border-clinic-600 text-clinic-600 transition';
      document.getElementById(tab.paneId)?.classList.remove('hidden');

      if (tab.paneId === 'tab-appointments') {
        fetchAppointments();
      } else if (tab.paneId === 'tab-practice') {
        fetchDoctorPracticeSettings();
      }
    });
  });
}

// ==========================================================================
// Doctor Directory & Available Slots
// ==========================================================================
async function fetchDoctors() {
  try {
    elements.doctorsList.innerHTML = `<div class="col-span-full py-8 text-center text-xs text-slate-400">Loading specialist directory...</div>`;
    const res = await apiRequest('/doctors');
    state.doctors = res.data || [];

    if (elements.statDoctorsCount) {
      elements.statDoctorsCount.textContent = `${state.doctors.length} Specialist${state.doctors.length === 1 ? '' : 's'}`;
    }

    if (state.doctors.length === 0) {
      elements.doctorsList.innerHTML = `<div class="col-span-full py-8 text-center text-xs text-slate-400">No doctors currently scheduled.</div>`;
      return;
    }

    renderDoctors();

    if (state.doctors.length > 0) {
      selectDoctor(state.doctors[0]);
    }
  } catch (err) {
    elements.doctorsList.innerHTML = `<div class="col-span-full py-8 text-center text-xs text-rose-500 font-medium">Failed to load specialists: ${err.message}</div>`;
  }
}

function renderDoctors() {
  elements.doctorsList.innerHTML = '';
  state.doctors.forEach((doc) => {
    const isSelected = state.selectedDoctor && state.selectedDoctor.id === doc.id;
    const card = document.createElement('div');

    card.className = `p-4 rounded-xl border transition-all cursor-pointer flex flex-col justify-between ${
      isSelected 
        ? 'border-clinic-600 bg-clinic-50/50 ring-2 ring-clinic-500/20 shadow-xs' 
        : 'border-slate-200 bg-white hover:border-clinic-300 hover:shadow-xs'
    }`;
    card.dataset.id = doc.id;

    const docName = doc.user ? doc.user.full_name : 'Specialist';
    const clinicName = doc.clinic ? doc.clinic.name : 'Central Clinic';

    card.innerHTML = `
      <div>
        <div class="flex items-center gap-3 mb-2.5">
          <div class="w-10 h-10 rounded-full bg-clinic-600 text-white font-bold text-sm flex items-center justify-center">
            ${docName.charAt(0)}
          </div>
          <div>
            <h4 class="text-sm font-bold text-slate-900 leading-tight">${docName}</h4>
            <span class="inline-block text-[11px] font-semibold text-clinic-700 bg-clinic-50 px-1.5 rounded border border-clinic-200 mt-0.5">
              ${doc.specialization}
            </span>
          </div>
        </div>
        <p class="text-xs text-slate-500 flex items-center gap-1 mb-3">
          <span>📍</span> <span>${clinicName}</span>
        </p>
      </div>

      <div class="pt-3 border-t border-slate-100 flex items-center justify-between text-xs">
        <div>
          <span class="text-slate-400 block text-[10px] uppercase font-semibold">Consultation</span>
          <span class="text-base font-extrabold text-emerald-700">$${Number(doc.consultation_fee).toFixed(2)}</span>
        </div>
        <span class="text-[11px] font-medium text-slate-500 bg-slate-100 px-2 py-1 rounded-md">
          ${doc.slot_duration_minutes} min visit
        </span>
      </div>
    `;

    card.addEventListener('click', () => {
      selectDoctor(doc);
      renderDoctors();
    });

    elements.doctorsList.appendChild(card);
  });
}

function selectDoctor(doctor) {
  state.selectedDoctor = doctor;
  state.selectedSlot = null;
  elements.sumDoctorName.textContent = doctor.user ? doctor.user.full_name : 'Physician';
  elements.sumSpecialization.textContent = doctor.specialization;
  elements.sumFee.textContent = `$${Number(doctor.consultation_fee).toFixed(2)}`;
  updateSummary();
  fetchAvailableSlots();
}

async function fetchAvailableSlots() {
  if (!state.selectedDoctor || !state.selectedDate) {
    elements.slotsGrid.innerHTML = `<div class="col-span-full py-6 text-center text-xs text-slate-400">Select a doctor and date to view available hours.</div>`;
    return;
  }

  elements.slotsGrid.innerHTML = `<div class="col-span-full py-6 text-center text-xs text-slate-400">Checking consultation schedule...</div>`;

  try {
    const res = await apiRequest(`/doctors/${state.selectedDoctor.id}/available-slots?date=${state.selectedDate}`);
    const slots = res.data.slots || [];

    if (slots.length === 0) {
      elements.slotsGrid.innerHTML = `<div class="col-span-full py-6 text-center text-xs text-slate-500">The doctor is not scheduled for consultations on this day.</div>`;
      return;
    }

    elements.slotsGrid.innerHTML = '';
    slots.forEach(slot => {
      const btn = document.createElement('button');
      btn.type = 'button';

      const isSelected = state.selectedSlot && state.selectedSlot.start_time === slot.start_time;

      if (isSelected) {
        btn.className = 'py-2.5 px-2 rounded-lg font-mono text-xs font-bold text-center bg-clinic-600 text-white border border-clinic-600 shadow-xs cursor-pointer ring-2 ring-clinic-500/30';
      } else if (slot.is_available) {
        btn.className = 'py-2.5 px-2 rounded-lg font-mono text-xs font-semibold text-center transition-all bg-emerald-50 text-emerald-700 border border-emerald-200 hover:bg-emerald-600 hover:text-white hover:border-emerald-600 cursor-pointer shadow-xs';
      } else {
        btn.className = 'py-2.5 px-2 rounded-lg font-mono text-xs font-semibold text-center bg-slate-100 text-slate-400 border border-slate-200 cursor-not-allowed line-through opacity-60';
        btn.disabled = true;
      }

      btn.textContent = slot.start_time.substring(0, 5);
      btn.title = `${slot.start_time.substring(0, 5)} - ${slot.end_time.substring(0, 5)} (${slot.is_available ? 'Available' : 'Booked'})`;

      if (slot.is_available) {
        btn.addEventListener('click', () => {
          state.selectedSlot = slot;
          updateSummary();
          fetchAvailableSlots();
        });
      }

      elements.slotsGrid.appendChild(btn);
    });
  } catch (err) {
    elements.slotsGrid.innerHTML = `<div class="col-span-full py-6 text-center text-xs text-rose-500">${err.message}</div>`;
  }
}

function updateSummary() {
  elements.sumDate.textContent = state.selectedDate || '-';
  elements.sumSlot.textContent = state.selectedSlot ? `${state.selectedSlot.start_time.substring(0, 5)} - ${state.selectedSlot.end_time.substring(0, 5)}` : '-';

  const canBook = state.selectedDoctor && state.selectedDate && state.selectedSlot && state.token;
  elements.btnConfirmBooking.disabled = !canBook;

  if (!state.token) {
    elements.btnConfirmBooking.innerHTML = `<span>Sign in as Patient to Book</span>`;
  } else {
    elements.btnConfirmBooking.innerHTML = `
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
        <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
      </svg>
      <span>Book Appointment</span>
    `;
  }
}

function getNextWeekdayString(daysAhead = 1) {
  const d = new Date();
  d.setDate(d.getDate() + daysAhead);
  while (d.getDay() === 0 || d.getDay() === 6) {
    d.setDate(d.getDate() + 1);
  }
  return d.toISOString().split('T')[0];
}

// ==========================================================================
// Booking Submission
// ==========================================================================
async function handleBookAppointment() {
  if (!state.token) {
    openAuthModal(false);
    return;
  }

  const complaint = elements.complaintText.value.trim();
  if (complaint.length < 5) {
    showToast('Please enter your reason for visit (minimum 5 characters).', 'error');
    elements.complaintText.focus();
    return;
  }

  elements.btnConfirmBooking.disabled = true;
  elements.btnConfirmBooking.textContent = 'Scheduling appointment...';

  try {
    const payload = {
      doctor_id: state.selectedDoctor.id,
      appointment_date: state.selectedDate,
      start_time: state.selectedSlot.start_time,
      complaint_description: complaint
    };

    await apiRequest('/appointments', {
      method: 'POST',
      body: JSON.stringify(payload)
    });

    showToast(`Your appointment has been scheduled for ${state.selectedSlot.start_time.substring(0, 5)}!`, 'success');
    elements.complaintText.value = '';
    
    await fetchAvailableSlots();
    
    setTimeout(() => {
      document.getElementById('tab-btn-appointments').click();
    }, 500);
  } catch (err) {
    if (err.status === 409 || err.errorCode === 'SLOT_ALREADY_BOOKED') {
      showToast('This time slot is no longer available. Please select another time.', 'error');
    } else {
      showToast(`Unable to complete booking: ${err.message}`, 'error');
    }
    await fetchAvailableSlots();
  } finally {
    updateSummary();
  }
}

// ==========================================================================
// Appointments Schedule & Patient Ledger
// ==========================================================================
async function fetchAppointments() {
  if (!elements.appointmentsList) return;

  if (!state.token) {
    elements.appointmentsList.innerHTML = `
      <div class="py-12 text-center text-xs text-slate-500">
        <p class="mb-3">Please sign in to view your scheduled visits.</p>
        <button class="px-3.5 py-1.5 text-xs font-semibold text-white bg-clinic-600 hover:bg-clinic-700 rounded-lg shadow-xs transition" onclick="document.getElementById('btn-quick-patient').click()">
          Sign In as Patient
        </button>
      </div>
    `;
    elements.badgeApptCount.textContent = '0';
    if (elements.statBookingsCount) elements.statBookingsCount.textContent = '0 Scheduled';
    return;
  }

  elements.appointmentsList.innerHTML = `<div class="py-12 text-center text-xs text-slate-400">Loading scheduled appointments...</div>`;

  try {
    const status = elements.filterStatus.value;
    const query = status ? `?status=${status}` : '';
    const res = await apiRequest(`/appointments/my${query}`);
    const appts = res.data || [];
    state.appointments = appts;
    elements.badgeApptCount.textContent = appts.length;
    if (elements.statBookingsCount) elements.statBookingsCount.textContent = `${appts.length} Scheduled`;

    if (appts.length === 0) {
      elements.appointmentsList.innerHTML = `<div class="py-12 text-center text-xs text-slate-400">No appointments found.</div>`;
      return;
    }

    let html = `
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-slate-200 text-xs text-left">
          <thead class="bg-slate-50 text-slate-500 font-semibold uppercase tracking-wider text-[10px]">
            <tr>
              <th scope="col" class="px-4 py-3">Date & Time</th>
              <th scope="col" class="px-4 py-3">Doctor</th>
              <th scope="col" class="px-4 py-3">Patient</th>
              <th scope="col" class="px-4 py-3">Reason for Visit</th>
              <th scope="col" class="px-4 py-3">Status</th>
              <th scope="col" class="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 bg-white">
    `;

    appts.forEach(appt => {
      const doctorName = appt.doctor && appt.doctor.user ? appt.doctor.user.full_name : 'Specialist';
      const patientName = appt.patient ? appt.patient.full_name : 'Patient';
      const userRole = state.user?.role;

      let statusBadge = 'bg-slate-100 text-slate-700 border-slate-200';
      if (appt.status === 'pending') statusBadge = 'bg-amber-50 text-amber-700 border-amber-200';
      if (appt.status === 'confirmed') statusBadge = 'bg-blue-50 text-blue-700 border-blue-200';
      if (appt.status === 'completed') statusBadge = 'bg-emerald-50 text-emerald-700 border-emerald-200';
      if (appt.status === 'cancelled') statusBadge = 'bg-rose-50 text-rose-700 border-rose-200';

      let actionsHtml = '';
      if (appt.status === 'pending') {
        if (userRole === 'doctor' || userRole === 'admin') {
          actionsHtml += `<button class="px-2.5 py-1 text-[11px] font-semibold text-white bg-clinic-600 hover:bg-clinic-700 rounded transition btn-confirm" data-id="${appt.id}">Confirm</button> `;
        }
        actionsHtml += `<button class="px-2.5 py-1 text-[11px] font-semibold text-slate-600 hover:text-rose-600 bg-slate-100 hover:bg-rose-50 rounded transition btn-cancel" data-id="${appt.id}">Cancel</button>`;
      } else if (appt.status === 'confirmed') {
        if (userRole === 'doctor') {
          actionsHtml += `<button class="px-2.5 py-1 text-[11px] font-semibold text-white bg-emerald-600 hover:bg-emerald-700 rounded transition btn-complete" data-id="${appt.id}">Complete Visit</button> `;
        }
        actionsHtml += `<button class="px-2.5 py-1 text-[11px] font-semibold text-slate-600 hover:text-rose-600 bg-slate-100 hover:bg-rose-50 rounded transition btn-cancel" data-id="${appt.id}">Cancel</button>`;
      }

      let notesHtml = `<span class="text-slate-600">"${appt.complaint_description}"</span>`;
      if (appt.medical_note) {
        notesHtml += `
          <div class="mt-1 p-2 rounded bg-emerald-50 border border-emerald-100 text-[11px]">
            <p class="font-bold text-emerald-800">Diagnosis: ${appt.medical_note.diagnosis}</p>
            ${appt.medical_note.prescription ? `<p class="text-emerald-700">Rx: ${appt.medical_note.prescription}</p>` : ''}
          </div>
        `;
      }

      html += `
        <tr class="hover:bg-slate-50/60 transition">
          <td class="px-4 py-3 whitespace-nowrap">
            <div class="font-semibold text-slate-900">${appt.appointment_date}</div>
            <div class="font-mono text-clinic-600 text-[11px]">${appt.start_time.substring(0, 5)} - ${appt.end_time.substring(0, 5)}</div>
          </td>
          <td class="px-4 py-3">
            <div class="font-bold text-slate-800">${doctorName}</div>
            <div class="text-slate-400 text-[11px]">${appt.doctor?.specialization || 'Cardiology'}</div>
          </td>
          <td class="px-4 py-3 whitespace-nowrap font-medium text-slate-700">
            ${patientName}
          </td>
          <td class="px-4 py-3 max-w-xs">
            ${notesHtml}
          </td>
          <td class="px-4 py-3 whitespace-nowrap">
            <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider border ${statusBadge}">
              ${appt.status}
            </span>
          </td>
          <td class="px-4 py-3 whitespace-nowrap text-right space-x-1">
            ${actionsHtml || '<span class="text-slate-400 text-[11px]">Completed</span>'}
          </td>
        </tr>
      `;
    });

    html += `</tbody></table></div>`;
    elements.appointmentsList.innerHTML = html;

    elements.appointmentsList.querySelectorAll('.btn-confirm').forEach(b => {
      b.addEventListener('click', () => handleConfirmAppt(b.dataset.id));
    });
    elements.appointmentsList.querySelectorAll('.btn-cancel').forEach(b => {
      b.addEventListener('click', () => handleCancelAppt(b.dataset.id));
    });
    elements.appointmentsList.querySelectorAll('.btn-complete').forEach(b => {
      b.addEventListener('click', () => openDiagnosisModal(b.dataset.id));
    });
  } catch (err) {
    elements.appointmentsList.innerHTML = `<div class="py-12 text-center text-xs text-rose-500">${err.message}</div>`;
  }
}

async function handleConfirmAppt(id) {
  try {
    await apiRequest(`/appointments/${id}/confirm`, { method: 'PATCH' });
    showToast('Appointment confirmed.', 'success');
    fetchAppointments();
  } catch (err) {
    showToast(`Error: ${err.message}`, 'error');
  }
}

async function handleCancelAppt(id) {
  const reason = prompt('Please enter a cancellation reason:', 'Schedule conflict');
  if (!reason) return;

  try {
    await apiRequest(`/appointments/${id}/cancel`, {
      method: 'PATCH',
      body: JSON.stringify({ reason })
    });
    showToast('Appointment cancelled.', 'info');
    fetchAppointments();
    fetchAvailableSlots();
  } catch (err) {
    showToast(`Cancellation error: ${err.message}`, 'error');
  }
}

function openDiagnosisModal(appointmentId) {
  elements.diagAppointmentId.value = appointmentId;
  elements.diagDiagnosis.value = '';
  elements.diagPrescription.value = '';
  elements.diagNotes.value = '';
  elements.modalDiagnosis.classList.remove('hidden');
}

async function handleCompleteAppointment(e) {
  e.preventDefault();
  const id = elements.diagAppointmentId.value;
  const payload = {
    diagnosis: elements.diagDiagnosis.value.trim(),
    prescription: elements.diagPrescription.value.trim(),
    notes: elements.diagNotes.value.trim()
  };

  try {
    await apiRequest(`/appointments/${id}/complete`, {
      method: 'POST',
      body: JSON.stringify(payload)
    });
    showToast('Consultation notes saved and visit marked complete.', 'success');
    elements.modalDiagnosis.classList.add('hidden');
    fetchAppointments();
  } catch (err) {
    showToast(`Unable to save notes: ${err.message}`, 'error');
  }
}

// ==========================================================================
// User Authentication Modal Handlers
// ==========================================================================
function openAuthModal(registerMode = false) {
  isRegisterMode = registerMode;
  elements.modalAuthTitle.textContent = isRegisterMode ? 'Register New Patient' : 'Sign In to ClinicEase';
  elements.groupFullname.classList.toggle('hidden', !isRegisterMode);
  elements.groupPhone.classList.toggle('hidden', !isRegisterMode);
  elements.btnSubmitAuth.textContent = isRegisterMode ? 'Register & Sign In' : 'Sign In';
  elements.btnToggleAuthMode.textContent = isRegisterMode 
    ? 'Already have an account? Sign In' 
    : "Don't have an account? Register Patient";
  elements.modalAuth.classList.remove('hidden');
}

async function handleAuthSubmit(e) {
  e.preventDefault();
  const email = elements.authEmail.value.trim();
  const password = elements.authPassword.value.trim();

  try {
    if (isRegisterMode) {
      const fullName = elements.authFullname.value.trim();
      const phone = elements.authPhone.value.trim();
      const res = await apiRequest('/auth/register', {
        method: 'POST',
        body: JSON.stringify({ email, password, full_name: fullName, phone })
      });
      setSession(res.data.token, res.data.user);
      showToast(`Welcome to ClinicEase, ${res.data.user.full_name}!`, 'success');
    } else {
      const res = await apiRequest('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password })
      });
      setSession(res.data.token, res.data.user);
      showToast(`Welcome back, ${res.data.user.full_name}!`, 'success');
    }
    elements.modalAuth.classList.add('hidden');
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// User Profile Switchers
async function loginAsPersona(email, password) {
  try {
    const res = await apiRequest('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });
    setSession(res.data.token, res.data.user);
    showToast(`Switched account to: ${res.data.user.full_name}`, 'success');
    return true;
  } catch (err) {
    showToast(`Unable to sign in as ${email}: ${err.message}`, 'error');
    return false;
  }
}

async function quickPatientLogin() {
  const patientEmail = 'patient@example.com';
  const patientPass = 'Password123!';
  
  try {
    const res = await apiRequest('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email: patientEmail, password: patientPass })
    });
    setSession(res.data.token, res.data.user);
    showToast(`Active account: ${res.data.user.full_name}`, 'success');
  } catch (err) {
    try {
      const reg = await apiRequest('/auth/register', {
        method: 'POST',
        body: JSON.stringify({
          email: patientEmail,
          password: patientPass,
          full_name: 'Jane Doe',
          phone: '+18005551234'
        })
      });
      setSession(reg.data.token, reg.data.user);
      showToast('Patient account created and signed in.', 'success');
    } catch (regErr) {
      showToast(`Unable to sign in: ${regErr.message}`, 'error');
    }
  }
}

async function handleAddDoctor(e) {
  e.preventDefault();

  let clinicId = state.doctors[0]?.clinic_id;
  if (!clinicId) {
    try {
      const clinicsRes = await apiRequest('/clinics');
      clinicId = clinicsRes.data[0]?.id;
    } catch (err) {
      showToast('No clinic facility found to assign specialist.', 'error');
      return;
    }
  }

  const payload = {
    full_name: document.getElementById('newdoc-name').value.trim(),
    specialization: document.getElementById('newdoc-specialization').value.trim(),
    email: document.getElementById('newdoc-email').value.trim(),
    password: document.getElementById('newdoc-password').value.trim(),
    consultation_fee: parseFloat(document.getElementById('newdoc-fee').value),
    slot_duration_minutes: parseInt(document.getElementById('newdoc-duration').value) || 30,
    phone: document.getElementById('newdoc-phone').value.trim(),
    bio: document.getElementById('newdoc-bio').value.trim(),
    clinic_id: clinicId
  };

  try {
    const res = await apiRequest('/admin/doctors', {
      method: 'POST',
      body: JSON.stringify(payload)
    });

    showToast(`Specialist ${res.data.user.full_name} created successfully!`, 'success');
    elements.modalAddDoctor.classList.add('hidden');
    elements.formAddDoctor.reset();
    fetchDoctors();
  } catch (err) {
    showToast(`Failed to add specialist: ${err.message}`, 'error');
  }
}

// ==========================================================================
// Doctor Practice Self-Service & Schedule Management
// ==========================================================================
const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

async function fetchDoctorPracticeSettings() {
  if (!state.token || !state.user || state.user.role !== 'doctor') return;

  try {
    const profileRes = await apiRequest('/doctors/me');
    const doc = profileRes.data;
    if (elements.docProfileName) elements.docProfileName.value = doc.user?.full_name || state.user.full_name;
    if (elements.docProfileSpecialization) elements.docProfileSpecialization.value = doc.specialization || '';
    if (elements.docProfileFee) elements.docProfileFee.value = doc.consultation_fee;
    if (elements.docProfileDuration) elements.docProfileDuration.value = doc.slot_duration_minutes;
    if (elements.docProfileBio) elements.docProfileBio.value = doc.bio || '';

    await fetchDoctorMySchedules();
  } catch (err) {
    showToast(`Could not load practice profile: ${err.message}`, 'error');
  }
}

async function fetchDoctorMySchedules() {
  if (!elements.doctorMySchedules) return;
  elements.doctorMySchedules.innerHTML = `<div class="py-4 text-center text-xs text-slate-400">Loading your schedule...</div>`;

  try {
    const res = await apiRequest('/doctors/me/schedules');
    const schedules = res.data || [];

    if (schedules.length === 0) {
      elements.doctorMySchedules.innerHTML = `
        <div class="p-4 rounded-lg bg-amber-50 border border-amber-200 text-amber-800 text-xs">
          No active consultation hours found. Add your available hours below so patients can book visits with you.
        </div>
      `;
      return;
    }

    elements.doctorMySchedules.innerHTML = '';
    // Sort by day of week
    schedules.sort((a, b) => a.day_of_week - b.day_of_week);

    schedules.forEach(sched => {
      const dayName = DAY_NAMES[sched.day_of_week] || `Day ${sched.day_of_week}`;
      const startTime = sched.start_time.substring(0, 5);
      const endTime = sched.end_time.substring(0, 5);

      const row = document.createElement('div');
      row.className = 'flex items-center justify-between p-3 rounded-lg border border-slate-200 bg-slate-50 text-xs hover:bg-white transition';
      row.innerHTML = `
        <div>
          <span class="font-bold text-slate-900 block">${dayName}</span>
          <span class="font-mono text-slate-500 text-[11px]">${startTime} – ${endTime}</span>
        </div>
        <button type="button" class="btn-delete-sched px-2 py-1 text-slate-400 hover:text-rose-600 hover:bg-rose-50 rounded text-xs font-semibold transition" data-id="${sched.id}">
          Remove
        </button>
      `;

      row.querySelector('.btn-delete-sched').addEventListener('click', () => handleDeleteSchedule(sched.id));
      elements.doctorMySchedules.appendChild(row);
    });
  } catch (err) {
    elements.doctorMySchedules.innerHTML = `<div class="py-4 text-center text-xs text-rose-500">${err.message}</div>`;
  }
}

async function handleSaveDoctorProfile(e) {
  e.preventDefault();
  const fee = parseFloat(elements.docProfileFee.value);
  const duration = parseInt(elements.docProfileDuration.value);
  const specialization = elements.docProfileSpecialization.value.trim();
  const bio = elements.docProfileBio.value.trim();

  if (isNaN(fee) || fee < 0) {
    showToast('Please enter a valid consultation fee.', 'error');
    return;
  }

  try {
    await apiRequest('/doctors/me', {
      method: 'PATCH',
      body: JSON.stringify({
        specialization,
        consultation_fee: fee,
        slot_duration_minutes: duration,
        bio
      })
    });

    showToast('Practice settings and consultation rates saved successfully!', 'success');
    fetchDoctors(); // Refresh public doctor directory
  } catch (err) {
    showToast(`Failed to update practice settings: ${err.message}`, 'error');
  }
}

async function handleAddDoctorSchedule(e) {
  e.preventDefault();
  const day = parseInt(elements.schedDay.value);
  let start = elements.schedStart.value;
  let end = elements.schedEnd.value;

  if (!start || !end) {
    showToast('Please specify both start and end times.', 'error');
    return;
  }

  if (start.length === 5) start += ':00';
  if (end.length === 5) end += ':00';

  try {
    await apiRequest('/doctors/me/schedules', {
      method: 'POST',
      body: JSON.stringify({
        day_of_week: day,
        start_time: start,
        end_time: end
      })
    });

    showToast('Consultation schedule slot published.', 'success');
    fetchDoctorMySchedules();
  } catch (err) {
    showToast(`Failed to add schedule: ${err.message}`, 'error');
  }
}

async function handleDeleteSchedule(id) {
  if (!confirm('Are you sure you want to remove this available consultation time slot?')) {
    return;
  }

  try {
    await apiRequest(`/doctors/me/schedules/${id}`, {
      method: 'DELETE'
    });

    showToast('Consultation hours removed.', 'info');
    fetchDoctorMySchedules();
  } catch (err) {
    showToast(`Failed to delete schedule: ${err.message}`, 'error');
  }
}

// ==========================================================================
// Initialization
// ==========================================================================
function init() {
  const defaultDate = getNextWeekdayString(1);
  state.selectedDate = defaultDate;
  elements.bookingDate.value = defaultDate;
  elements.bookingDate.min = new Date().toISOString().split('T')[0];

  elements.btnOpenLogin.addEventListener('click', () => openAuthModal(false));
  elements.btnOpenRegister.addEventListener('click', () => openAuthModal(true));
  elements.btnCloseAuth.addEventListener('click', () => elements.modalAuth.classList.add('hidden'));
  elements.btnLogout.addEventListener('click', logout);
  elements.formAuth.addEventListener('submit', handleAuthSubmit);

  elements.btnToggleAuthMode.addEventListener('click', () => {
    openAuthModal(!isRegisterMode);
  });

  elements.btnQuickDoctor.addEventListener('click', () => loginAsPersona('doctor@clinicease.com', 'Doctor123!'));
  elements.btnQuickAdmin.addEventListener('click', () => loginAsPersona('admin@clinicease.com', 'Admin123!'));
  elements.btnQuickPatient.addEventListener('click', quickPatientLogin);

  elements.bookingDate.addEventListener('change', (e) => {
    state.selectedDate = e.target.value;
    updateSummary();
    fetchAvailableSlots();
  });

  elements.btnDateTomorrow.addEventListener('click', () => {
    const d = getNextWeekdayString(1);
    elements.bookingDate.value = d;
    state.selectedDate = d;
    updateSummary();
    fetchAvailableSlots();
  });

  elements.btnDateMonday.addEventListener('click', () => {
    const d = getNextWeekdayString(3);
    elements.bookingDate.value = d;
    state.selectedDate = d;
    updateSummary();
    fetchAvailableSlots();
  });

  elements.btnConfirmBooking.addEventListener('click', handleBookAppointment);
  elements.btnRefreshAppts.addEventListener('click', fetchAppointments);
  elements.filterStatus.addEventListener('change', fetchAppointments);

  elements.formDiagnosis.addEventListener('submit', handleCompleteAppointment);
  elements.btnCloseDiagnosis.addEventListener('click', () => elements.modalDiagnosis.classList.add('hidden'));
  elements.btnCancelDiagnosis.addEventListener('click', () => elements.modalDiagnosis.classList.add('hidden'));

  // Admin Add Doctor events
  if (elements.btnOpenAddDoctor) {
    elements.btnOpenAddDoctor.addEventListener('click', () => {
      elements.modalAddDoctor.classList.remove('hidden');
    });
  }
  if (elements.btnCloseAddDoctor) {
    elements.btnCloseAddDoctor.addEventListener('click', () => {
      elements.modalAddDoctor.classList.add('hidden');
    });
  }
  if (elements.btnCancelAddDoctor) {
    elements.btnCancelAddDoctor.addEventListener('click', () => {
      elements.modalAddDoctor.classList.add('hidden');
    });
  }
  if (elements.formAddDoctor) {
    elements.formAddDoctor.addEventListener('submit', handleAddDoctor);
  }

  // Doctor self-service events
  if (elements.formDoctorProfile) {
    elements.formDoctorProfile.addEventListener('submit', handleSaveDoctorProfile);
  }
  if (elements.formAddSchedule) {
    elements.formAddSchedule.addEventListener('submit', handleAddDoctorSchedule);
  }

  initTabs();
  updateAuthUI();
  fetchDoctors();
  verifyProfile();
}

document.addEventListener('DOMContentLoaded', init);
