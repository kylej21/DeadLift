# ── VPC ──────────────────────────────────────────────────────────────────────

resource "google_compute_network" "vpc" {
  name                    = "deadlift-vpc"
  auto_create_subnetworks = false
  depends_on              = [google_project_service.apis]
}

# ── Classic VPN ───────────────────────────────────────────────────────────────
# Connects GCP → on-prem machine at 10.30.112.192

resource "google_compute_address" "vpn_ip" {
  name   = "deadlift-vpn-ip"
  region = var.region
}

resource "google_compute_vpn_gateway" "vpn_gw" {
  name    = "deadlift-vpn-gw"
  network = google_compute_network.vpc.id
  region  = var.region
}

# Classic VPN requires three forwarding rules to handle IKE + ESP traffic.
resource "google_compute_forwarding_rule" "vpn_esp" {
  name        = "deadlift-vpn-esp"
  region      = var.region
  ip_protocol = "ESP"
  ip_address  = google_compute_address.vpn_ip.address
  target      = google_compute_vpn_gateway.vpn_gw.self_link
}

resource "google_compute_forwarding_rule" "vpn_udp500" {
  name        = "deadlift-vpn-udp500"
  region      = var.region
  ip_protocol = "UDP"
  port_range  = "500"
  ip_address  = google_compute_address.vpn_ip.address
  target      = google_compute_vpn_gateway.vpn_gw.self_link
}

resource "google_compute_forwarding_rule" "vpn_udp4500" {
  name        = "deadlift-vpn-udp4500"
  region      = var.region
  ip_protocol = "UDP"
  port_range  = "4500"
  ip_address  = google_compute_address.vpn_ip.address
  target      = google_compute_vpn_gateway.vpn_gw.self_link
}

resource "google_compute_vpn_tunnel" "onprem" {
  name               = "deadlift-onprem-tunnel"
  region             = var.region
  peer_ip            = var.onprem_vpn_peer_ip
  shared_secret      = var.onprem_vpn_shared_secret
  target_vpn_gateway = google_compute_vpn_gateway.vpn_gw.id

  # Accept traffic from/to the full on-prem CIDR through the tunnel.
  local_traffic_selector  = ["0.0.0.0/0"]
  remote_traffic_selector = [var.onprem_cidr]

  depends_on = [
    google_compute_forwarding_rule.vpn_esp,
    google_compute_forwarding_rule.vpn_udp500,
    google_compute_forwarding_rule.vpn_udp4500,
  ]
}

# Static route: send traffic destined for the on-prem machine through the tunnel.
resource "google_compute_route" "onprem" {
  name                = "deadlift-onprem-route"
  network             = google_compute_network.vpc.name
  dest_range          = "10.40.115.4/32"
  next_hop_vpn_tunnel = google_compute_vpn_tunnel.onprem.id
  priority            = 1000
}

# ── Firewall rules ────────────────────────────────────────────────────────────
# Allow return traffic from the on-prem network back into the VPC (responses
# to connections initiated by Cloud Run through the connector).
resource "google_compute_firewall" "allow_onprem_ingress" {
  name    = "deadlift-allow-onprem-ingress"
  network = google_compute_network.vpc.name

  direction     = "INGRESS"
  source_ranges = [var.onprem_cidr]

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "icmp"
  }
}

# Allow egress from the VPC connector range to the on-prem network (belt-and-
# suspenders — GCP allows all egress by default, but this makes intent explicit).
resource "google_compute_firewall" "allow_onprem_egress" {
  name    = "deadlift-allow-onprem-egress"
  network = google_compute_network.vpc.name

  direction          = "EGRESS"
  destination_ranges = [var.onprem_cidr]

  allow {
    protocol = "tcp"
  }

  allow {
    protocol = "icmp"
  }
}

# ── VPC Access Connector ──────────────────────────────────────────────────────
# Gives Cloud Run egress into the VPC (and therefore the VPN tunnel).

resource "google_vpc_access_connector" "connector" {
  name          = "deadlift-connector"
  region        = var.region
  network       = google_compute_network.vpc.name
  ip_cidr_range = "10.8.0.0/28" # /28 reserved exclusively for the connector
  machine_type  = "e2-micro"
  min_instances = 2
  max_instances = 3

  depends_on = [google_project_service.apis]
}
