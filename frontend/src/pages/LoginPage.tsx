import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate, Link } from 'react-router-dom'
import { z } from 'zod'
import { login } from '../api/auth'
import { useAuthStore } from '../store/authStore'
import './AuthPages.css'

const loginSchema = z.object({
  username: z.string().min(3, '请输入用户名或邮箱'),
  password: z.string().min(6, '密码至少6位'),
})

type LoginForm = z.infer<typeof loginSchema>

const LoginPage = () => {
  const navigate = useNavigate()
  const loginStore = useAuthStore((state) => state.login)
  const [error, setError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (values: LoginForm) => {
    setError(null)
    try {
      const result = await login(values)
      loginStore(result)
      navigate('/images')
    } catch (err: any) {
      setError(err.response?.data?.message ?? '登录失败')
    }
  }

  return (
    <div className="auth-wrapper">
      <form className="auth-card" onSubmit={handleSubmit(onSubmit)}>
        <h2>登录</h2>
        {error && <div className="error-banner">{error}</div>}

        <label>
          用户名或邮箱
          <input type="text" {...register('username')} />
          {errors.username && <span className="field-error">{errors.username.message}</span>}
        </label>

        <label>
          密码
          <input type="password" {...register('password')} />
          {errors.password && <span className="field-error">{errors.password.message}</span>}
        </label>

        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? '登录中...' : '登录'}
        </button>

        <p className="switch-link">
          还没有账号？<Link to="/register">立即注册</Link>
        </p>
      </form>
    </div>
  )
}

export default LoginPage

