import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { MemoryRouter } from "react-router"
import { describe, expect, it } from "vitest"
import PasswordResetConfirmPage from "@/pages/password-reset-confirm-page"

describe("PasswordResetConfirmPage", () => {
  it("新しいパスワードを確認欄に合わせて直すと、不一致のエラーが消える", async () => {
    const user = userEvent.setup()
    render(
      <MemoryRouter initialEntries={["/password-reset/confirm?token=t"]}>
        <PasswordResetConfirmPage />
      </MemoryRouter>,
    )
    const newPassword = screen.getByLabelText("新しいパスワード")
    await user.type(newPassword, "abcdefgh")
    await user.type(screen.getByLabelText("パスワード確認"), "abcdefgX")
    await user.click(screen.getByRole("button", { name: /パスワード/ }))
    expect(await screen.findByText("パスワードが一致しません")).toBeInTheDocument()

    await user.clear(newPassword)
    await user.type(newPassword, "abcdefgX")

    expect(screen.queryByText("パスワードが一致しません")).not.toBeInTheDocument()
  })
})
