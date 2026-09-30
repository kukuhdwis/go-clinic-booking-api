# Business Interface & Anti-AI Slop Guidelines

## Core Principles
1. **Never Expose Internal Technical Architecture to End-Users**:
   - Do NOT display backend implementation jargon in user interfaces (e.g., "Atomic Locking", "Pessimistic Concurrency", "Zero Double-Booking Guarantee", "SELECT ... FOR UPDATE", "Microservices Engine", "Database Transactions").
   - These are internal backend engineering specifications, not business or patient-facing copy.

2. **Authentic Domain Language**:
   - Always communicate in natural, professional business and healthcare terms:
     - Healthcare: "Book an Appointment", "Specialist Directory", "Consultation Hours", "Reason for Visit", "Clinical Notes & Diagnosis", "Electronic Prescription".
     - Error messages: "This time slot is no longer available. Please select another time." instead of "409 CONFLICT: Pessimistic Lock Row Acquired".

3. **Production Standard UX**:
   - Build interfaces matching real-world enterprise standards (like Epic, Cerner, Doctolib, Stripe, Linear).
   - Metrics cards must show real domain KPIs (e.g. "Available Doctors", "Today's Appointments", "Clinic Hours"), never technical architecture metrics.
   - Developer tools (inspectors, simulated race condition testers) must NEVER clutter the primary business workflows.
