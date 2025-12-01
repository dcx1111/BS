import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useNavigate } from 'react-router-dom'
import { z } from 'zod'
import { register as registerApi, login } from '../api/auth'
import { useAuthStore } from '../store/authStore'
import './AuthPages.css'

const registerSchema = z.object({
  username: z.string().min(6, '用户名至少6个字符'),
  email: z.string().email('请输入正确的邮箱'),
  password: z.string().min(6, '密码至少6位'),
})

type RegisterForm = z.infer<typeof registerSchema>

const RegisterPage = () => {
  const navigate = useNavigate()
  const loginStore = useAuthStore((state) => state.login)
  const [error, setError] = useState<string | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  })

  const onSubmit = async (values: RegisterForm) => {
    setError(null)
    try {
      await registerApi(values)
      const result = await login({ username: values.username, password: values.password })
      loginStore(result)
      navigate('/images')
    } catch (err: any) {
      setError(err.response?.data?.message ?? '注册失败')
    }
  }

  return (
    <div className="auth-wrapper">
      <form className="auth-card" onSubmit={handleSubmit(onSubmit)}>
        <h2>注册</h2>
        {error && <div className="error-banner">{error}</div>}

        <label>
          用户名
          <input type="text" {...register('username')} />
          {errors.username && <span className="field-error">{errors.username.message}</span>}
        </label>

        <label>
          邮箱
          <input type="email" {...register('email')} />
          {errors.email && <span className="field-error">{errors.email.message}</span>}
        </label>

        <label>
          密码
          <input type="password" {...register('password')} />
          {errors.password && <span className="field-error">{errors.password.message}</span>}
        </label>

        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? '提交中...' : '创建账号'}
        </button>

        <p className="switch-link">
          已有账号？<Link to="/login">去登录</Link>
        </p>
      </form>
    </div>
  )
}

export default RegisterPage

