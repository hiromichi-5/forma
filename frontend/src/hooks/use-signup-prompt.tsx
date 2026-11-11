import { useNavigate } from "react-router"
import { toast } from "sonner"
import { Button } from "@/components/ui/button"

/** デモで保存を伴う操作をしたときに、アカウント作成へ案内する。 */
export function useSignupPrompt() {
  const navigate = useNavigate()

  return () =>
    toast.custom((id) => (
      <div className="flex w-[356px] items-center gap-3 rounded-2xl bg-white p-4 shadow-lg">
        <div className="flex-1 text-sm">
          <p className="font-medium text-foreground">デモでは変更を保存できません</p>
          <p className="mt-0.5 text-xs text-muted-foreground">アカウントを作成すると使えます</p>
        </div>
        <Button
          size="sm"
          className="rounded-full px-4"
          onClick={() => {
            toast.dismiss(id)
            void navigate("/signup")
          }}
        >
          無料ではじめる
        </Button>
      </div>
    ))
}
