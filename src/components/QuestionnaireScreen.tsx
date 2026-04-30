import React, { useState, useRef } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Question, UserResponse } from '../types/personality';
import { ProgressBar } from './ProgressBar';

interface QuestionnaireScreenProps {
  questions: Question[];
  onComplete: (responses: UserResponse[]) => void;
  nickname: string;
}

export const QuestionnaireScreen: React.FC<QuestionnaireScreenProps> = ({ 
  questions, 
  onComplete, 
  nickname 
}) => {
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [responses, setResponses] = useState<UserResponse[]>([]);
  const isCompletedRef = useRef(false);

  const handleAnswer = (value: number) => {
    if (isCompletedRef.current) return;

    const newResponse: UserResponse = {
      questionId: questions[currentQuestion].id,
      value
    };

    const updatedResponses = responses.filter(r => r.questionId !== newResponse.questionId);
    updatedResponses.push(newResponse);
    setResponses(updatedResponses);

    // Lock immediately on final question to prevent double-submit
    if (currentQuestion === questions.length - 1) {
      isCompletedRef.current = true;
    }

    // Auto-advance to next question
    setTimeout(() => {
      if (currentQuestion < questions.length - 1) {
        setCurrentQuestion(currentQuestion + 1);
      } else {
        onComplete(updatedResponses);
      }
    }, 500);
  };

  const goToPrevious = () => {
    if (currentQuestion > 0) {
      setCurrentQuestion(currentQuestion - 1);
    }
  };

  const getCurrentResponse = () => {
    return responses.find(r => r.questionId === questions[currentQuestion].id)?.value;
  };

  const responseLabels = [
    'Strongly Disagree',
    'Disagree',
    'Neutral',
    'Agree',
    'Strongly Agree'
  ];

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="max-w-4xl w-full">
        {/* Header */}
        <div className="text-center mb-8">
          <h2 className="text-2xl font-bold text-gray-800 mb-2">
            Hey {nickname}! 👋
          </h2>
          <p className="text-gray-600">
            Question {currentQuestion + 1} of {questions.length}
          </p>
          <ProgressBar 
            current={currentQuestion + 1} 
            total={questions.length} 
            className="mt-4 max-w-md mx-auto"
          />
        </div>

        {/* Question Card */}
        <div className="bg-white/90 backdrop-blur-sm rounded-3xl p-8 shadow-2xl mb-8 animate-slide-up">
          <div className="text-center mb-8">
            <h3 className="text-3xl font-bold text-gray-800 mb-4 leading-tight">
              {questions[currentQuestion].text}
            </h3>
            <p className="text-gray-500">
              How much do you agree with this statement?
            </p>
          </div>

          {/* Response Options */}
          <div className="space-y-3">
            {responseLabels.map((label, index) => {
              const value = index + 1;
              const isSelected = getCurrentResponse() === value;
              
              return (
                <button
                  key={value}
                  onClick={() => handleAnswer(value)}
                  className={`w-full p-4 rounded-xl text-left font-medium transition-all duration-300 transform hover:scale-105 ${
                    isSelected
                      ? 'bg-gradient-to-r from-purple-500 to-pink-500 text-white shadow-lg'
                      : 'bg-gray-50 hover:bg-gray-100 text-gray-700'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span>{label}</span>
                    <div className={`w-4 h-4 rounded-full border-2 ${
                      isSelected
                        ? 'bg-white border-white'
                        : 'border-gray-300'
                    }`} />
                  </div>
                </button>
              );
            })}
          </div>
        </div>

        {/* Navigation */}
        <div className="flex justify-between items-center">
          <button
            onClick={goToPrevious}
            disabled={currentQuestion === 0}
            className={`flex items-center space-x-2 px-6 py-3 rounded-xl font-medium transition-all duration-300 ${
              currentQuestion === 0
                ? 'text-gray-400 cursor-not-allowed'
                : 'text-purple-600 hover:bg-purple-50'
            }`}
          >
            <ChevronLeft className="w-5 h-5" />
            <span>Previous</span>
          </button>

          <div className="text-sm text-gray-500">
            {Math.round(((currentQuestion + 1) / questions.length) * 100)}% Complete
          </div>

          <div className="w-24"></div> {/* Spacer for alignment */}
        </div>
      </div>
    </div>
  );
};