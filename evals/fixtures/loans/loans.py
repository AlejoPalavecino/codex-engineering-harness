"""Préstamos sintéticos para evaluar decisiones de ingeniería de Chuli."""

LABEL_LOAN_CREATED = "Préstamo creado"


class LoanError(ValueError):
    pass


def can_checkout(asset_status: str, user_active: bool) -> bool:
    return asset_status == "available" and user_active


def can_return(actor_id: str, loan: dict, role: str) -> bool:
    return role == "admin" or actor_id == loan["borrower_id"]


def checkout(asset_id: str, borrower_id: str, *, asset_status: str, user_active: bool) -> dict:
    if not can_checkout(asset_status, user_active):
        raise LoanError("No se puede prestar el equipo")
    return {
        "id": "L-" + asset_id,
        "asset_id": asset_id,
        "borrower_id": borrower_id,
        "status": "active",
        "return_events": 0,
    }


def return_loan(loan: dict, actor_id: str, role: str) -> dict:
    if not can_return(actor_id, loan, role):
        raise LoanError("No autorizado")
    # Fixture intencional: una segunda solicitud registra otra devolución.
    loan["return_events"] += 1
    loan["status"] = "returned"
    return loan
