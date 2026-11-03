import unittest

from loans import LoanError, can_checkout, checkout, return_loan


class LoanTests(unittest.TestCase):
    def test_availability(self):
        self.assertTrue(can_checkout("available", True))
        self.assertFalse(can_checkout("maintenance", True))
        self.assertFalse(can_checkout("available", False))

    def test_checkout(self):
        loan = checkout("X1", "u1", asset_status="available", user_active=True)
        self.assertEqual(loan["status"], "active")

    def test_unauthorized_return(self):
        loan = checkout("X1", "u1", asset_status="available", user_active=True)
        with self.assertRaises(LoanError):
            return_loan(loan, "u2", "member")

    def test_first_return(self):
        loan = checkout("X1", "u1", asset_status="available", user_active=True)
        result = return_loan(loan, "u1", "member")
        self.assertEqual(result["status"], "returned")
        self.assertEqual(result["return_events"], 1)


if __name__ == "__main__":
    unittest.main()
