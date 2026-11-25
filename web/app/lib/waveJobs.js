// The jobs the Wave family is built for, by setting (2026-11-24). One source for the job finder in
// the manual's Appendix B5 (WaveJobs.svelte) and the per-slot summary on research-wave-family.html §3.
// Moved verbatim from the family page's table when it went to the manual (founder, 2026-11-24).
// `slot` is written for people ("Micro to Giga", "Pico + Nano", "Nano, Micro"); wave-jobs.js and the
// guards read it the same way. `scope: "beyond-plant"` marks the rows the research hub's scope reads.
export const LADDER = ["Pico", "Nano", "Micro", "Giga", "Tera", "Peta", "Exa"];

export const SETTINGS = [
  {
    "key": "plant",
    "label": "Plant and process",
    "sub": "oil and gas, power, water, manufacturing",
    "jobs": [
      {
        "name": "Gas leak screening",
        "slot": "Pico + Nano",
        "io": "Sensor and acoustic stream to anomaly event",
        "why": "Continuous, no bandwidth, hazardous area"
      },
      {
        "name": "Leak survey to repair order",
        "slot": "Micro",
        "io": "Findings to repair work order",
        "why": "Emissions reporting from a disconnected site"
      },
      {
        "name": "Rotating-equipment health",
        "slot": "Nano to Giga",
        "io": "Vibration and thermal features to alert and cause",
        "why": "Latency, connectivity, data volume"
      },
      {
        "name": "Alarm triage and flood reduction",
        "slot": "Micro to Giga",
        "io": "Alarm stream to ranked, deduplicated summary",
        "why": "The control network cannot send data out"
      },
      {
        "name": "Work-order and log automation",
        "slot": "Micro",
        "io": "Free-text notes to structured record",
        "why": "The task measured to survive heavy compression"
      },
      {
        "name": "Procedure and drawing questions",
        "slot": "Micro + Giga",
        "io": "Question and local corpus to grounded answer",
        "why": "Documents never leave the plant"
      },
      {
        "name": "Incident, permit, and handover notes",
        "slot": "Giga",
        "io": "Notes to structured report and summary",
        "why": "Privacy, audit trail, disconnected network"
      },
      {
        "name": "Inspection and root cause",
        "slot": "Micro to Giga",
        "io": "Image and line data to defect call with reason",
        "why": "Line-rate latency"
      },
      {
        "name": "Gloves-on voice control",
        "slot": "Pico to Nano",
        "io": "Spoken command in plant noise to a confirmed action request",
        "why": "Hands and eyes are busy, and the area may be hazardous"
      },
      {
        "name": "Radio and intercom transcription",
        "slot": "Nano to Micro",
        "io": "Channel audio to a searchable shift log",
        "why": "Continuous audio, no bandwidth to ship it"
      },
      {
        "name": "Gauge and panel reading",
        "slot": "Micro",
        "io": "Photograph of a dial or screen to a numeric reading",
        "why": "Legacy assets with no digital output"
      },
      {
        "name": "Cause codes from technician notes",
        "slot": "Nano to Micro",
        "io": "Free-text repair note to a standard failure code",
        "why": "Notes stay on site; codes feed the local historian"
      },
      {
        "name": "Spare-part and BOM lookup",
        "slot": "Nano + Micro",
        "io": "Part described in words to a catalog number",
        "why": "The catalog is local and the site is disconnected"
      },
      {
        "name": "Procedure step-out detection",
        "slot": "Micro",
        "io": "Executed steps against the written procedure to a deviation flag",
        "why": "Controlled procedures never leave the plant"
      },
      {
        "name": "Sensor drift and calibration triage",
        "slot": "Nano to Giga",
        "io": "Instrument trend to suspected drift and its likely cause",
        "why": "Continuous data, thin link, fast turnaround"
      },
      {
        "name": "Shift handover summary",
        "slot": "Micro to Giga",
        "io": "A shift of logs and alarms to what the next crew must know",
        "why": "Personnel and incident detail stays on site"
      },
      {
        "name": "Emissions and regulatory reporting",
        "slot": "Giga",
        "io": "Field findings to a submission-ready record",
        "why": "Audit trail, and the site is disconnected"
      },
      {
        "name": "Weld and assembly inspection",
        "slot": "Micro to Giga",
        "io": "Line image to accept or reject, with the defect named",
        "why": "Line-rate latency; images never leave the cell"
      }
    ]
  },
  {
    "key": "clinical",
    "label": "Hospital and lab equipment",
    "sub": "the equipment and facilities layer, never the patient",
    "jobs": [
      {
        "name": "Imaging fleet health",
        "slot": "Micro to Giga",
        "io": "Scanner service log to fault and likely cause",
        "why": "Vendor-segmented network, nothing leaves the hospital"
      },
      {
        "name": "Biomedical work orders",
        "slot": "Nano to Micro",
        "io": "Technician note to structured equipment failure record",
        "why": "Records stay on site, beside the patient network"
      },
      {
        "name": "Alarm configuration review",
        "slot": "Micro",
        "io": "Alarm journal to misconfigured limits and duplicate annunciations",
        "why": "Journals never leave the clinical network"
      },
      {
        "name": "Device reportability routing",
        "slot": "Micro to Giga",
        "io": "Device event to reporting trigger, citation, and deadline",
        "why": "Regulated record, retained on premises"
      },
      {
        "name": "Cold-chain excursions",
        "slot": "Pico to Nano",
        "io": "Refrigeration telemetry to an excursion event",
        "why": "Continuous, no bandwidth, unattended overnight"
      },
      {
        "name": "Sterile processing records",
        "slot": "Nano to Micro",
        "io": "Sterilizer cycle record to a compliance exception",
        "why": "High volume, and the record stays with the department"
      }
    ]
  },
  {
    "key": "defence",
    "label": "Defence sustainment",
    "sub": "keeping the fleet serviceable, no effects",
    "jobs": [
      {
        "name": "Platform fault triage",
        "slot": "Micro to Giga",
        "io": "Maintenance log to a coded fault and its likely cause",
        "why": "Denied and intermittent links, no reachback"
      },
      {
        "name": "Readiness paperwork",
        "slot": "Nano to Micro",
        "io": "Free-text status to a structured readiness record",
        "why": "Air-gapped network"
      },
      {
        "name": "Supply catalog lookup",
        "slot": "Nano",
        "io": "Part described in words to catalog candidates",
        "why": "Disconnected depot, local catalog"
      }
    ]
  },
  {
    "key": "workshop",
    "label": "Workshop, home and small shop",
    "sub": "one Pi-class board on a bench, and up",
    "jobs": [
      {
        "name": "Bench-machine watch",
        "slot": "Pico",
        "io": "Vibration and motor current from a lathe, mill or printer to a stop-now event",
        "why": "One board on the bench, and no uplink worth paying for",
        "scope": "beyond-plant"
      },
      {
        "name": "Print-farm failure detection",
        "slot": "Pico to Nano",
        "io": "Camera frames and motion trace to a failed-print call",
        "why": "The frames never leave the workshop",
        "scope": "beyond-plant"
      },
      {
        "name": "Kiln and heat-treat cycle log",
        "slot": "Pico",
        "io": "Thermocouple trend to a clean, held or overshot cycle",
        "why": "It runs unattended overnight, on its own board",
        "scope": "beyond-plant"
      },
      {
        "name": "Off-grid power health",
        "slot": "Pico to Nano",
        "io": "Battery, charger and inverter telemetry to a state-of-health call",
        "why": "A boat, a van or a cabin, with no link to spare",
        "scope": "beyond-plant"
      },
      {
        "name": "Tool crib and parts-bin lookup",
        "slot": "Nano",
        "io": "A part described in words to the bin it lives in",
        "why": "The shop’s own catalogue, kept offline",
        "scope": "beyond-plant"
      },
      {
        "name": "Job cards from bench notes",
        "slot": "Nano to Micro",
        "io": "A scribbled note to a job card the shop can search",
        "why": "Customer names stay on the shop’s own machine",
        "scope": "beyond-plant"
      },
      {
        "name": "Home sensor mesh roll-up",
        "slot": "Nano",
        "io": "A dozen small sensor nodes to one plain state per zone",
        "why": "Nothing leaves the house",
        "scope": "beyond-plant"
      },
      {
        "name": "Setup check before a cut",
        "slot": "Micro",
        "io": "A job file and the machine’s limits to a go or no-go, with the reason",
        "why": "One desktop GPU in the corner, not a subscription",
        "scope": "beyond-plant"
      }
    ]
  },
  {
    "key": "fleet",
    "label": "Across sites and fleets",
    "sub": "where the top of the ladder earns its size",
    "jobs": [
      {
        "name": "Reliability report for leadership",
        "slot": "Tera",
        "io": "A quarter of plant records to a written report, with every figure it cites",
        "why": "The numbers have to tie back to the estate’s own record",
        "scope": "beyond-plant"
      },
      {
        "name": "Standard operating procedure drafting",
        "slot": "Tera",
        "io": "Incident history and the vendor manual to a draft SOP an engineer edits",
        "why": "Procedures encode how this company runs, not how the industry does",
        "scope": "beyond-plant"
      },
      {
        "name": "Vendor and contractor performance review",
        "slot": "Tera",
        "io": "Work orders, callbacks and downtime by supplier to a review pack",
        "why": "Supplier terms and performance are commercially confidential",
        "scope": "beyond-plant"
      },
      {
        "name": "Turnaround and outage scope package",
        "slot": "Tera to Peta",
        "io": "Deferred work, inspection findings and history to a scoped, sequenced work list",
        "why": "The plan says what is broken and when the plant will be down",
        "scope": "beyond-plant"
      },
      {
        "name": "Capital request justification",
        "slot": "Tera",
        "io": "Failure and downtime history to a business case, with the cost of doing nothing",
        "why": "It is built out of the company’s own loss record",
        "scope": "beyond-plant"
      },
      {
        "name": "Incident investigation across sites",
        "slot": "Tera",
        "io": "One failure and every similar event in the estate to a root-cause report",
        "why": "The evidence is operating data from every site it touches",
        "scope": "beyond-plant"
      },
      {
        "name": "Regulatory filings across an estate",
        "slot": "Tera",
        "io": "Monitoring records to a drafted filing with its evidence attached",
        "why": "The filing goes in under the operator’s own name",
        "scope": "beyond-plant"
      },
      {
        "name": "Operator training material",
        "slot": "Tera",
        "io": "The estate’s procedures and its real incidents to lesson material",
        "why": "It teaches this plant rather than a generic one",
        "scope": "beyond-plant"
      },
      {
        "name": "Regional reliability benchmarking",
        "slot": "Peta",
        "io": "Many sites’ records to a like-for-like comparison and the outliers",
        "why": "Ranking your own sites is an internal conversation",
        "scope": "beyond-plant"
      },
      {
        "name": "Multi-year capital planning",
        "slot": "Peta",
        "io": "Condition, age and failure history across a region to a funding sequence",
        "why": "The sequence is the region’s strategy",
        "scope": "beyond-plant"
      },
      {
        "name": "Standardisation proposals",
        "slot": "Peta",
        "io": "How each site does the same task to one harmonised procedure, and what each site must change",
        "why": "It puts every site’s real practice side by side",
        "scope": "beyond-plant"
      },
      {
        "name": "Spares and inventory policy",
        "slot": "Peta",
        "io": "Failure rates and lead times across a region to stocking levels per site",
        "why": "Stock levels show where an operator is exposed",
        "scope": "beyond-plant"
      },
      {
        "name": "Scenario and risk modelling",
        "slot": "Peta to Exa",
        "io": "A proposed change against a decade of process history to what it does by year ten",
        "why": "The estate’s own model of itself, built on its own record",
        "scope": "beyond-plant"
      },
      {
        "name": "Engineering standards from the whole record",
        "slot": "Exa",
        "io": "Every procedure, incident and spec in the estate to one standard, and the conflicts it found",
        "why": "It reads the whole corpus at once, and the corpus is the company",
        "scope": "beyond-plant"
      },
      {
        "name": "The case nothing below could close",
        "slot": "Exa",
        "io": "A failure the smaller slots left open, with everything each of them saw attached",
        "why": "The escalation carries the raw evidence with it",
        "scope": "beyond-plant"
      },
      {
        "name": "Teaching the smaller tiers",
        "slot": "Exa",
        "io": "The flagship’s working on hard cases to training signal for the slots below",
        "why": "The teaching material is the company’s own operating history",
        "scope": "beyond-plant"
      },
      {
        "name": "A new site with no history",
        "slot": "Exa",
        "io": "A plant with no record of its own to a working model, drawn from the fleet’s",
        "why": "The fleet record is the asset, and it never leaves to make one",
        "scope": "beyond-plant"
      }
    ]
  }
];

/** the smallest slot a job names: the first rung of the ladder its slot cell mentions */
export function smallest(slot) {
  return LADDER.find((s) => new RegExp(`\\b${s}\\b`).test(slot));
}
