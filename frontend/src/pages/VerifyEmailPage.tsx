import { useEffect, useRef, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { Card, CardContent, CardTitle } from "../components/ui/card";
import { apiClient } from "../lib/api";
import { getApiErrorMessage } from "../lib/api-error";
import { CheckCircle, Loader2, XCircle } from "lucide-react";

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const [status, setStatus] = useState<"loading" | "success" | "error">("loading");
  const [errorMessage, setErrorMessage] = useState("");
  const verifiedRef = useRef(false);

  useEffect(() => {
    if (!token) {
      setStatus("error");
      setErrorMessage("認証トークンが見つかりません");
      return;
    }

    if (verifiedRef.current) {
      setStatus("success");
      return;
    }

    const verify = async () => {
      try {
        await apiClient.verifyEmail({ token });
        verifiedRef.current = true;
        setStatus("success");
      } catch (err) {
        if (verifiedRef.current) {
          setStatus("success");
          return;
        }
        setStatus("error");
        setErrorMessage(
          getApiErrorMessage(
            err,
            {
              TOKEN_NOT_FOUND: "トークンが無効または期限切れです",
              VALIDATION_ERROR: "認証トークンが不正です",
              NETWORK_ERROR: "ネットワークエラーが発生しました",
            },
            "認証に失敗しました"
          )
        );
      }
    };

    verify();
  }, [token]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted p-4">
      <Card className="w-full max-w-md shadow-lg">
        <CardContent className="p-6">
          <div className="space-y-6 text-center">
            {status === "loading" && (
              <>
                <Loader2 className="h-12 w-12 animate-spin text-primary mx-auto" />
                <CardTitle className="text-xl">メールアドレスを確認中...</CardTitle>
              </>
            )}

            {status === "success" && (
              <>
                <div className="flex justify-center">
                  <div className="rounded-full bg-green-100 p-4">
                    <CheckCircle className="h-12 w-12 text-green-600" />
                  </div>
                </div>
                <div>
                  <CardTitle className="text-xl">メール認証が完了しました</CardTitle>
                  <p className="mt-2 text-sm text-muted-foreground">
                    アカウントが有効化されました。ログインしてご利用ください。
                  </p>
                </div>
                <Link
                  to="/login"
                  className="inline-block font-medium text-primary hover:underline"
                >
                  ログインページへ
                </Link>
              </>
            )}

            {status === "error" && (
              <>
                <div className="flex justify-center">
                  <div className="rounded-full bg-destructive/10 p-4">
                    <XCircle className="h-12 w-12 text-destructive" />
                  </div>
                </div>
                <div>
                  <CardTitle className="text-xl">認証に失敗しました</CardTitle>
                  <p className="mt-2 text-sm text-muted-foreground">{errorMessage}</p>
                </div>
                <Link
                  to="/login"
                  className="inline-block font-medium text-primary hover:underline"
                >
                  ログインページへ
                </Link>
              </>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
