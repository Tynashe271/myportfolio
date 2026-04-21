from django.shortcuts import render
from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_http_methods
import json

def portfolio_view(request):
    """Main portfolio view"""
    context = {
        'name': 'Tinashe Nyenyesa',
        'badge': 'Industrial Attachment Candidate · July 2026',
        'description': 'Second-year Computer Science student at NUST (Part 2.2) with a strong passion for IoT, cyber security, and full-stack development. Certified in Cyber Security. Seeking an industrial attachment from July 2026 to contribute and grow with experienced teams.',
        'phone': '+263 77 959 5732',
        'phone_raw': '+263779595732',
        'email_university': 'tinashenyenyesa@students.nust.ac.zw',
        'email_personal': 'tinashenyenyesa10@gmail.com',
        'github': 'https://github.com/Tynashe271',
        'github_handle': 'Tynashe271',
        'certification': 'Professional Certificate of Proficiency in Cyber Security',
        'cert_issuer': 'Lupane State University – Centre for Continuing Education',
        'cert_date': 'Jan-Mar 2026',
        'cert_credential': 'LSU-CCE02690',
        'languages': 'English (fluent) · Shona (native)',
        'availability': 'July 2026 – April 2027',
        'interests': 'IoT, Software Engineering, Cyber Security, Embedded Systems',
        'location': 'Bulawayo, Zimbabwe · Open to remote/hybrid opportunities',
        'learning': 'Advanced JavaScript, Node.js, Cloud basics',
        'skills': [
            {'name': 'Python', 'icon': 'fab fa-python'},
            {'name': 'Java', 'icon': 'fab fa-java'},
            {'name': 'JavaScript', 'icon': 'fab fa-js'},
            {'name': 'SQL', 'icon': 'fas fa-database'},
            {'name': 'HTML/CSS', 'icon': 'fab fa-html5'},
            {'name': 'Git & VS Code', 'icon': 'fab fa-git-alt'},
            {'name': 'Arduino IDE', 'icon': 'fas fa-microchip'},
            {'name': 'Kali Linux', 'icon': 'fas fa-shield-alt'},
            {'name': 'Data Comms', 'icon': 'fas fa-network-wired'},
        ],
        'projects': [
            {
                'title': 'SmartBin – IoT Waste Management System',
                'description': 'Sensor-based bin monitoring system using ultrasonic technology. Learning to implement automated alerts for collection teams to optimize waste logistics.',
                'status': 'In progress · 2025-2026',
                'status_icon': 'fas fa-code-branch',
                'icon': 'fas fa-trash-alt fa-2x',
                'icon_secondary': 'fas fa-microchip',
            },
            {
                'title': 'Smart Health Monitoring System',
                'description': 'Developing a wearable-enabled remote patient monitoring solution. Integrating sensor data with a web dashboard for real-time health metrics.',
                'status': 'In progress · 2025-2026',
                'status_icon': 'fas fa-microchip',
                'icon': 'fas fa-heartbeat fa-2x',
                'icon_secondary': 'fas fa-chart-line',
            },
            {
                'title': 'Personal Portfolio Website',
                'description': 'Responsive portfolio built with HTML/CSS/JS, hosted on GitHub Pages. Practicing version control, responsive design, and front-end deployment workflows.',
                'status': 'Live · 2025-Present',
                'status_icon': 'fab fa-github',
                'icon': 'fas fa-laptop-code fa-2x',
                'icon_secondary': '',
            },
        ],
    }
    return render(request, 'portfolio/index.html', context)


@csrf_exempt
@require_http_methods(["POST"])
def contact_email(request):
    """Handle contact form submission via email"""
    try:
        data = json.loads(request.body)
        name = data.get('name', '').strip()
        email = data.get('email', '').strip()
        message = data.get('message', '').strip()
        
        if not name or not email or not message:
            return JsonResponse({'success': False, 'error': 'Please fill all fields.'}, status=400)
        
        # For development, just return success
        # In production, you would send an email here
        
        return JsonResponse({'success': True, 'message': 'Message sent successfully!'})
        
    except Exception as e:
        return JsonResponse({'success': False, 'error': str(e)}, status=500)